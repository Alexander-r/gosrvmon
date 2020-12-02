CREATE TABLE public.hosts
(
  host text NOT NULL,
  CONSTRAINT hosts_host_pkey PRIMARY KEY (host)
);

CREATE TABLE public.checks
(
  host text NOT NULL,
  check_time timestamp without time zone NOT NULL,
  rtt bigint NOT NULL,
  up boolean NOT NULL,
  CONSTRAINT checks_pkey PRIMARY KEY (host, check_time),
  CONSTRAINT checks_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (host) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX checks_check_time_idx
  ON public.checks
  USING brin
  (host, check_time);

CREATE TABLE public.state_change_params
(
  host text NOT NULL,
  change_threshold bigint NOT NULL,
  action text NOT NULL,
  CONSTRAINT state_change_params_pkey PRIMARY KEY (host),
  CONSTRAINT state_change_params_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (host) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX state_change_params_host_idx
  ON public.state_change_params
  USING brin
  (host);
