PGDMP      
                }            testplatform    17.4    17.4     �           0    0    ENCODING    ENCODING        SET client_encoding = 'UTF8';
                           false            �           0    0 
   STDSTRINGS 
   STDSTRINGS     (   SET standard_conforming_strings = 'on';
                           false            �           0    0 
   SEARCHPATH 
   SEARCHPATH     8   SELECT pg_catalog.set_config('search_path', '', false);
                           false            �           1262    16387    testplatform    DATABASE     r   CREATE DATABASE testplatform WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'ru-RU';
    DROP DATABASE testplatform;
                     postgres    false            �            1259    16407    answers    TABLE     �   CREATE TABLE public.answers (
    id uuid NOT NULL,
    question_id uuid,
    text text NOT NULL,
    correct boolean DEFAULT false NOT NULL
);
    DROP TABLE public.answers;
       public         heap r       postgres    false            �            1259    16395 	   questions    TABLE     b   CREATE TABLE public.questions (
    id uuid NOT NULL,
    test_id uuid,
    text text NOT NULL
);
    DROP TABLE public.questions;
       public         heap r       postgres    false            �            1259    16388    tests    TABLE     M   CREATE TABLE public.tests (
    id uuid NOT NULL,
    title text NOT NULL
);
    DROP TABLE public.tests;
       public         heap r       postgres    false            �            1259    16577    user_test_results    TABLE       CREATE TABLE public.user_test_results (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    test_id uuid,
    correct_answers integer NOT NULL,
    total_questions integer NOT NULL,
    submitted_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);
 %   DROP TABLE public.user_test_results;
       public         heap r       postgres    false            �            1259    16420    users    TABLE     �  CREATE TABLE public.users (
    id uuid NOT NULL,
    telegram_id bigint NOT NULL,
    is_admin boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    first_name text DEFAULT ''::text NOT NULL,
    last_name text DEFAULT ''::text NOT NULL,
    group_name text DEFAULT ''::text NOT NULL
);
    DROP TABLE public.users;
       public         heap r       postgres    false            �          0    16407    answers 
   TABLE DATA           A   COPY public.answers (id, question_id, text, correct) FROM stdin;
    public               postgres    false    219   �       �          0    16395 	   questions 
   TABLE DATA           6   COPY public.questions (id, test_id, text) FROM stdin;
    public               postgres    false    218          �          0    16388    tests 
   TABLE DATA           *   COPY public.tests (id, title) FROM stdin;
    public               postgres    false    217   /       �          0    16577    user_test_results 
   TABLE DATA           q   COPY public.user_test_results (id, user_id, test_id, correct_answers, total_questions, submitted_at) FROM stdin;
    public               postgres    false    221   L       �          0    16420    users 
   TABLE DATA           u   COPY public.users (id, telegram_id, is_admin, created_at, updated_at, first_name, last_name, group_name) FROM stdin;
    public               postgres    false    220   i       >           2606    16414    answers answers_pkey 
   CONSTRAINT     R   ALTER TABLE ONLY public.answers
    ADD CONSTRAINT answers_pkey PRIMARY KEY (id);
 >   ALTER TABLE ONLY public.answers DROP CONSTRAINT answers_pkey;
       public                 postgres    false    219            <           2606    16401    questions questions_pkey 
   CONSTRAINT     V   ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_pkey PRIMARY KEY (id);
 B   ALTER TABLE ONLY public.questions DROP CONSTRAINT questions_pkey;
       public                 postgres    false    218            :           2606    16394    tests tests_pkey 
   CONSTRAINT     N   ALTER TABLE ONLY public.tests
    ADD CONSTRAINT tests_pkey PRIMARY KEY (id);
 :   ALTER TABLE ONLY public.tests DROP CONSTRAINT tests_pkey;
       public                 postgres    false    217            D           2606    16583 (   user_test_results user_test_results_pkey 
   CONSTRAINT     f   ALTER TABLE ONLY public.user_test_results
    ADD CONSTRAINT user_test_results_pkey PRIMARY KEY (id);
 R   ALTER TABLE ONLY public.user_test_results DROP CONSTRAINT user_test_results_pkey;
       public                 postgres    false    221            @           2606    16427    users users_pkey 
   CONSTRAINT     N   ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
 :   ALTER TABLE ONLY public.users DROP CONSTRAINT users_pkey;
       public                 postgres    false    220            B           2606    16429    users users_telegram_id_key 
   CONSTRAINT     ]   ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_telegram_id_key UNIQUE (telegram_id);
 E   ALTER TABLE ONLY public.users DROP CONSTRAINT users_telegram_id_key;
       public                 postgres    false    220            F           2606    16415     answers answers_question_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.answers
    ADD CONSTRAINT answers_question_id_fkey FOREIGN KEY (question_id) REFERENCES public.questions(id) ON DELETE CASCADE;
 J   ALTER TABLE ONLY public.answers DROP CONSTRAINT answers_question_id_fkey;
       public               postgres    false    219    218    4668            E           2606    16402     questions questions_test_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.questions
    ADD CONSTRAINT questions_test_id_fkey FOREIGN KEY (test_id) REFERENCES public.tests(id) ON DELETE CASCADE;
 J   ALTER TABLE ONLY public.questions DROP CONSTRAINT questions_test_id_fkey;
       public               postgres    false    217    4666    218            G           2606    16589 0   user_test_results user_test_results_test_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.user_test_results
    ADD CONSTRAINT user_test_results_test_id_fkey FOREIGN KEY (test_id) REFERENCES public.tests(id);
 Z   ALTER TABLE ONLY public.user_test_results DROP CONSTRAINT user_test_results_test_id_fkey;
       public               postgres    false    221    217    4666            H           2606    16584 0   user_test_results user_test_results_user_id_fkey    FK CONSTRAINT     �   ALTER TABLE ONLY public.user_test_results
    ADD CONSTRAINT user_test_results_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
 Z   ALTER TABLE ONLY public.user_test_results DROP CONSTRAINT user_test_results_user_id_fkey;
       public               postgres    false    220    221    4672            �      x������ � �      �      x������ � �      �      x������ � �      �      x������ � �      �      x������ � �     