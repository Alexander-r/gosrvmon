CREATE TABLE public.hosts
(
  id SERIAL NOT NULL,
  host text NOT NULL,
  UNIQUE(id),
  CONSTRAINT hosts_host_pkey PRIMARY KEY (host)
);

CREATE TABLE public.checks
(
  host integer NOT NULL,
  check_time timestamp without time zone NOT NULL,
  rtt bigint NOT NULL,
  up boolean NOT NULL,
  CONSTRAINT checks_pkey PRIMARY KEY (host, check_time),
  CONSTRAINT checks_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX checks_check_time_idx
  ON public.checks
  USING brin
  (host, check_time);

CREATE TABLE public.notifications_params
(
  host integer NOT NULL,
  change_threshold bigint NOT NULL,
  action text NOT NULL,
  CONSTRAINT notifications_params_pkey PRIMARY KEY (host),
  CONSTRAINT notifications_params_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX notifications_params_host_idx
  ON public.notifications_params
  USING brin
  (host);
