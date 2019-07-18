--
-- PostgreSQL database dump
--

-- Dumped from database version 9.6.13
-- Dumped by pg_dump version 11.3 (Debian 11.3-1.pgdg90+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: messages; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.messages (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    thread_id integer NOT NULL,
    sent_by_id integer NOT NULL,
    content character varying(4096) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.messages OWNER TO handcarry;

--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.messages_id_seq OWNER TO handcarry;

--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.messages_id_seq OWNED BY public.messages.id;


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.organizations (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    url character varying(255),
    auth_type character varying(255) NOT NULL,
    auth_config jsonb NOT NULL,
    uuid uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.organizations OWNER TO handcarry;

--
-- Name: organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.organizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.organizations_id_seq OWNER TO handcarry;

--
-- Name: organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.organizations_id_seq OWNED BY public.organizations.id;


--
-- Name: posts; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.posts (
    id integer NOT NULL,
    created_by_id integer NOT NULL,
    type character varying(255) NOT NULL,
    org_id integer NOT NULL,
    status character varying(255) NOT NULL,
    title character varying(255) NOT NULL,
    destination character varying(255),
    origin character varying(255),
    size character varying(255) NOT NULL,
    uuid uuid NOT NULL,
    receiver_id integer,
    provider_id integer,
    needed_after date NOT NULL,
    needed_before date NOT NULL,
    category character varying(255) NOT NULL,
    description character varying(4096),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.posts OWNER TO handcarry;

--
-- Name: posts_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.posts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.posts_id_seq OWNER TO handcarry;

--
-- Name: posts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.posts_id_seq OWNED BY public.posts.id;


--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO handcarry;

--
-- Name: thread_participants; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.thread_participants (
    id integer NOT NULL,
    thread_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.thread_participants OWNER TO handcarry;

--
-- Name: thread_participants_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.thread_participants_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.thread_participants_id_seq OWNER TO handcarry;

--
-- Name: thread_participants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.thread_participants_id_seq OWNED BY public.thread_participants.id;


--
-- Name: threads; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.threads (
    id integer NOT NULL,
    uuid uuid NOT NULL,
    post_id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.threads OWNER TO handcarry;

--
-- Name: threads_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.threads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.threads_id_seq OWNER TO handcarry;

--
-- Name: threads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.threads_id_seq OWNED BY public.threads.id;


--
-- Name: user_access_tokens; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.user_access_tokens (
    id integer NOT NULL,
    user_id integer NOT NULL,
    access_token character varying(255) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_access_tokens OWNER TO handcarry;

--
-- Name: user_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.user_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_access_tokens_id_seq OWNER TO handcarry;

--
-- Name: user_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.user_access_tokens_id_seq OWNED BY public.user_access_tokens.id;


--
-- Name: user_organizations; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.user_organizations (
    id integer NOT NULL,
    org_id integer NOT NULL,
    user_id integer NOT NULL,
    role character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.user_organizations OWNER TO handcarry;

--
-- Name: user_organizations_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.user_organizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_organizations_id_seq OWNER TO handcarry;

--
-- Name: user_organizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.user_organizations_id_seq OWNED BY public.user_organizations.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: handcarry
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    nickname character varying(255) NOT NULL,
    auth_org_id integer NOT NULL,
    auth_org_uid character varying(255) NOT NULL,
    admin_role character varying(255),
    uuid uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.users OWNER TO handcarry;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: handcarry
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO handcarry;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: handcarry
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.messages ALTER COLUMN id SET DEFAULT nextval('public.messages_id_seq'::regclass);


--
-- Name: organizations id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.organizations ALTER COLUMN id SET DEFAULT nextval('public.organizations_id_seq'::regclass);


--
-- Name: posts id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.posts ALTER COLUMN id SET DEFAULT nextval('public.posts_id_seq'::regclass);


--
-- Name: thread_participants id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.thread_participants ALTER COLUMN id SET DEFAULT nextval('public.thread_participants_id_seq'::regclass);


--
-- Name: threads id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.threads ALTER COLUMN id SET DEFAULT nextval('public.threads_id_seq'::regclass);


--
-- Name: user_access_tokens id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.user_access_tokens_id_seq'::regclass);


--
-- Name: user_organizations id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations ALTER COLUMN id SET DEFAULT nextval('public.user_organizations_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: posts posts_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_pkey PRIMARY KEY (id);


--
-- Name: thread_participants thread_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.thread_participants
    ADD CONSTRAINT thread_participants_pkey PRIMARY KEY (id);


--
-- Name: threads threads_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.threads
    ADD CONSTRAINT threads_pkey PRIMARY KEY (id);


--
-- Name: user_access_tokens user_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: user_organizations user_organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: messages_uuid_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX messages_uuid_idx ON public.messages USING btree (uuid);


--
-- Name: organizations_uuid_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX organizations_uuid_idx ON public.organizations USING btree (uuid);


--
-- Name: posts_uuid_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX posts_uuid_idx ON public.posts USING btree (uuid);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- Name: threads_uuid_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX threads_uuid_idx ON public.threads USING btree (uuid);


--
-- Name: user_access_tokens_access_token_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX user_access_tokens_access_token_idx ON public.user_access_tokens USING btree (access_token);


--
-- Name: users_auth_org_id_auth_org_uid_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX users_auth_org_id_auth_org_uid_idx ON public.users USING btree (auth_org_id, auth_org_uid);


--
-- Name: users_email_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX users_email_idx ON public.users USING btree (email);


--
-- Name: users_nickname_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX users_nickname_idx ON public.users USING btree (nickname);


--
-- Name: users_uuid_idx; Type: INDEX; Schema: public; Owner: handcarry
--

CREATE UNIQUE INDEX users_uuid_idx ON public.users USING btree (uuid);


--
-- Name: messages messages_sent_by_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_sent_by_id_fkey FOREIGN KEY (sent_by_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: messages messages_thread_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.threads(id) ON DELETE CASCADE;


--
-- Name: posts posts_created_by_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_created_by_id_fkey FOREIGN KEY (created_by_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: posts posts_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: posts posts_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: posts posts_receiver_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.posts
    ADD CONSTRAINT posts_receiver_id_fkey FOREIGN KEY (receiver_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: thread_participants thread_participants_thread_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.thread_participants
    ADD CONSTRAINT thread_participants_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.threads(id) ON DELETE CASCADE;


--
-- Name: thread_participants thread_participants_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.thread_participants
    ADD CONSTRAINT thread_participants_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: threads threads_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.threads
    ADD CONSTRAINT threads_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id) ON DELETE CASCADE;


--
-- Name: user_access_tokens user_access_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_access_tokens
    ADD CONSTRAINT user_access_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE;


--
-- Name: user_organizations user_organizations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.user_organizations
    ADD CONSTRAINT user_organizations_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: users users_auth_org_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: handcarry
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_auth_org_id_fkey FOREIGN KEY (auth_org_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--

