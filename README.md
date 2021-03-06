# Gosrvmon

Gosrvmon is a free, open-source, self-hosted uptime monitoring system.

Gosrvmon periodically performs server availability checks using ICMP (ping), TCP or HTTP/HTTPS requests. Check results and response time are stored in database. Results can be presented as SVG chart or as raw json data. All data can be accessed via simple web interface or using POST or GET requests. Results from multiple monitoring instances can be combined into a s single chart.

Chart example:
![gosrvmon chart example](chart.svg)

## Setup

Gosrvmon can be installed by checking out the repository and using `go build` command.
Or you can use `go get github.com/Alexander-r/gosrvmon` command to download the source code and build it. The resulting binary will be located at ```$GOPATH/bin/gosrvmon```.

Binary can be downloaded from GitLab mirror from pipeline artifacts (download artifacts button on the right next to the pipeline) at:
https://gitlab.com/Sonnix/gosrvmon/-/pipelines

ICMP (ping) checks require permission to create raw socket. This can be done by running application as root or by granting permission using capabilities:

```
setcap cap_net_raw+ep gosrvmon
```

Without any parameters default values will be used. Configuration can be loaded from file using `-config` option and providing path for configuration file:

```
gosrvmon -config /etc/gosrvmon.json
```

A sample configuration can be found in [config.json](config.json) file.

Configuration can also be provided as a string from command line using `-confstr` option and a json configuration string:

```
gosrvmon -confstr '{"DB":{"Type":"bolt","Database":"gosrvmon.db"}}'
```

or

```
gosrvmon -confstr "{\"DB\":{\"Type\":\"bolt\",\"Database\":\"gosrvmon.db\"}}"
```

Only one of this options should be used at a time.

Any value can be omitted from configuration. Default value will be used in it's place.

Without specifying Database parameter Gosrvmon will store everything in RAM and all data will be lost on restart. This mode is designed for testing purposes.

For persistent storage Gosrvmon requires a database. It can use PostgreSQL or an embedded database (Bolt DB or ql) which will write the data to a single file.
PostgreSQL provides good performance and more flexible data access and management.
Bolt DB is recommended if you don't want to use external database.
ql is used for in memory storage. It can also be used for persistent storage but will provide worse performance on large amounts of data.

On 32 bit systems embedded database may be limited to 2Gb storage depending on how much virtual memory can be addressed by mmap. You can limit the data stored using retention settings or use PostgreSQL or an external database if you require more storage.

PostgreSQL or embedded database require structure initialization. In-memory database will be initialized automatically. If a file for embedded database does not exist then it will be created and initialized. 

PostgreSQL database can be initialized by running  gosrvmon with `-init` option.

```
gosrvmon -config /etc/gosrvmon.json -init
```

PostgreSQL can also be initialized manually using [init.sql](init.sql).

## Adding hosts for monitoring

Host can be added by domain name or by IP address. Check method is selected based on how a host is added for monitoring:

 * `http://example.org/` would result in HTTP check.
 * `example.org:80` would result in TCP check.
 * `example.org` would result in ICMP check.

IPv6 hosts are also supported (for example `http://[2606:2800:220:1:248:1893:25c8:1946]\`, `[2606:2800:220:1:248:1893:25c8:1946]:80` , `2606:2800:220:1:248:1893:25c8:1946`). If host is added by domain name which has multiple A and AAAA records and ICMP check method is used then the request will be sent to every address and host is considered online if any of the addresses sends the response.

For HTTP checks host is considered available only if 2XX or 3XX response code was received. Any other response code (such as 404 or 401) will be considered as server being offline.

## Configuration

### DB
 * `Type` - can be `"pg"` or `"pq"` for PostgreSQL, `"bolt"` for Bolt, `"ql"` for ql embedded database. Default value is `"bolt"`.
 * `Host` - host for PostgreSQL connection.
 * `Port` - port for PostgreSQL connection.
 * `User` - user name for PostgreSQL connection.
 * `Password` - password for PostgreSQL connection.
 * `Database` - database name for PostgreSQL connection or path to database file for Bolt or ql embedded database. If left blank `""` then in-memory database will be used.
 
### Listen
 * `Address` - address on which the embedded web server should listen. Can be left blank `""` for listening an all available interfaces.
 * `Port` - port on which the embedded web server should listen. Default value is `8000`.
 * `ReadTimeout` - read timeout for embedded web server (in seconds). Default value is `30`.
 * `WriteTimeout` - write timeout for embedded web server (in seconds). Default value is `60`.

#### WebAuth
 * `Enable` - if enabled actions like adding or removing hosts would require basic http authentication. Default value is `false` (no authentication).
 * `User` - user name for basic http authentication.
 * `Password` - password for basic http authentication.

For better authentication control it is advised to set up a reverse proxy web server and use external authentication services.
 
### Checks
 * `Timeout` - timeout after which the host is considered to be offline (in seconds). Default value is `10`.
 * `Interval` - how often the checks should be performed (in seconds). Default value is `60`.
 * `PingRetryCount` - number of ping attempts for ICMP check. Default value is `4`.
 * `HTTPMethod` - which http method to use in requests. Can be `"GET"` for standard GET requests of `"HEAD"` for requesting only page headers. Default value is `"GET"`.
 * `PerformChecks` - if enabled periodic checks will be performed. When disabled the application will not perform any checks and will only serve historic data or display data aggregated from other instances. Default value is `true`.
 * `UseRemoteChecks` - if enabled application will request additional checks data from remote servers. If multiple servers monitor the same host then in the resulting chart the host will be considered online if at least one server was able to connect to it. If multiple servers were able to connect to the host then the lowest latency will be displayed. Default value is `false`.
 * `RemoteChecksURLs` - an array of servers from which additional data will be requested. Multiple servers can be set like this : `[ "http://192.168.1.1:8000/api/checks", "http://192.168.1.2:8000/api/checks" , "http://192.168.1.3:8000/api/checks" ]`
 * `AllowSingleChecks` - if enables single checks of host current state can be performed. The result of this check will be presented as json data or in web interface and will not be stored to database. Default value is `false`.
 * `Retention` - retention period for historic data (in seconds). Any data older than this value will periodically removed from database to free space. If set to `0` than no periodic cleanups will be performed and all data will be stored for as long as there is free space. Default value is `0`.

### Chart
 * `MaxRttScale` - Maximum timeout value for chart Y scale (in milliseconds). Default value is `200`.
 * `DynamicRttScale` - if enabled a minimal required timeout value for chart Y scale would be used up to MaxRttScale. If disabled then the scale will always go up to MaxRttScale.
 * `TimeZone` - time zone name in which to display dates to user on charts. Time zone name can be `UTC` for UTC, `Local` for local time or a location name from  IANA Time Zone database (for example `America/New_York`). Default value is `UTC`.

## Notifications

Gosrvmon supports sending notification when host changes state (goes offline or online).

Notifications can be set up at ```/web/notifications_params``` endpoint.

Host should be already set up at ```/web/hosts``` endpoint. It should be entered here exactly as it was entered at ```/web/hosts``` including protocol and port.

Change threshold is the amount of consecutive checks with a new state after which the notification will be sent. For example if you set it to 3 and the host goes offline then the notification will be sent after 3 consecutive checks which show that the host is now offline.
This setting may be useful if you are using ping check as it is not always reliable and can sometimes fail. So if you get a random packet loss it will not trigger the notification for that check.

Action is an HTTP or HTTPS URL that will be accessed to send the notification. Currently it only supports GET requests.

For example to send the notification to Telegram using bot API action can be set to something like this:

```https://api.telegram.org/bot<token>/sendMessage?chat_id=<chat_id>&text=State%20changed.%20Host%3A%20{HOST}%20New%20state%3A%20{STATE}%20Time%3A%20{TIME}```

```<token>``` should be replaced with bot access token. ```<chat_id>``` should be replaced with chat ID. The message should be URL encoded for the request to work.
```{HOST}``` will be replaced with the host that triggered the notification. ```{STATE}``` will be replaced by the hosts new state. {TIME} will be replaced by the string containing the time of the event. This parts should not be URL encoded as they are processed before the request is sent.

This will result a notification with this message:

```State changed. Host: 8.8.8.8 New state: down Time: 2021-03-02 12:00:00 +0000 UTC```

You can set it up to work with any other messaging service that provides similar bot API or send e-mail or an SMS message if you have a a gateway that can send messages using an HTTP API call.

In the message you can use the following placeholder strings that will be replaced before the API call with relevant data:

 * ```{HOST}``` - the host that triggered the notification.
 * ```{TIME}``` - a string with time and date of the event.
 * ```{TIMESTAMP}``` - unix timestamp of the event.
 * ```{RTT}``` - rtt of the event that triggered the notification. Will be an actual rtt if the host went online. Will be a timeout if the host went offline.
 * ```{RTTSTR}``` - rtt as a string.
 * ```{STATE}``` - will be up if the host went online or down if the host went offline.
 * ```{UP}``` - will be true if the host is online and false if the host is offline.
 * ```{DOWN}``` - will be false if the host is online and true if the host is offline.

## Backup and Restore

Gosrvmon can export hosts list and notification parameters as a json file. You can get the file using GET request on `/api/backup` endpoint:

```
curl http://127.0.0.1:8000/api/backup --output backup.json
```

You can restore the backup by sending this file using POST request to this endpoint:

```
curl -H 'Content-Type: application/json' -X POST http://127.0.0.1:8000/api/backup -d @backup.json
```

If enabled authentication may be required to access this endpoint:

```
curl -H 'Content-Type: application/json' -X POST http://127.0.0.1:8000/api/backup -d @backup.json -u user:password
```

To perform a full backup you can use same requests with `/api/backup_full` endpoint. This will also save checks results for all hosts.
This backup may be large in size and will require a lot of memory to process depending on how much is stored.

```
curl http://127.0.0.1:8000/api/backup_full --output backup_full.json
```

```
curl -H 'Content-Type: application/json' -X POST http://127.0.0.1:8000/api/backup_full -d @backup_full.json
```

Backups can be used to migrate from one database type to another. It can also be used between updates when internal data structure is changed.

## Docker

Gosrvmon is also provided as a Docker container:

```
docker pull sonnix1/gosrvmon
```

To run the container a valid configuration is required. Database structure needs to be initialized as described in setup section.

To initialize database using docker container run:

```
docker run -v /opt/gosrvmon/:/opt/gosrvmon/ sonnix1/gosrvmon -config /opt/gosrvmon/config.json -init
```

Run the docker container:

```
docker run -d -v /opt/gosrvmon/:/opt/gosrvmon/ -p 8000:8000 sonnix1/gosrvmon -config /opt/gosrvmon/config.json
```

Don't forget to provide a valid volume with configuration file and optionally database file if embedded database is used.

Container has default configuration file located at `/config.json`. It can be used for testing using in-memory database. This does not require additional initialization or setup but the data will reset on container restart. To use this configuration file run:

```
docker run -d -p 8000:8000 sonnix1/gosrvmon -config /config.json
```
