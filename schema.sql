create table authors
(
    id         INTEGER
        primary key autoincrement,
    name       TEXT,
    created_at DATETIME default CURRENT_TIMESTAMP not null,
    updated_at DATETIME default CURRENT_TIMESTAMP not null
);

create table books
(
    id          INTEGER
        primary key autoincrement,
    name        TEXT,
    edition     TEXT,
    description TEXT,
    created_at  DATETIME default CURRENT_TIMESTAMP not null,
    updated_at  DATETIME default CURRENT_TIMESTAMP not null
);

create table book_author
(
    book_id   integer
        constraint fk_book_author_book
            references books,
    author_id integer
        constraint fk_book_author_author
            references authors,
    primary key (book_id, author_id)
);

create table sqlite_master
(
    type     text,
    name     text,
    tbl_name text,
    rootpage int,
    sql      text
);

create table sqlite_sequence
(
    name,
    seq
);

create table user
(
    id                INTEGER
        constraint user_pk
            primary key autoincrement,
    email             text not null,
    created_at        datetime default current_timestamp,
    updated_at        datetime default current_timestamp,
    profile_image_url text,
    name              text not null,
    remember_token    TEXT     default 'xxx' not null
);

create table collection
(
    id         INTEGER
        primary key autoincrement,
    user_id    INTEGER not null
        references user,
    name       TEXT    not null,
    created_at DATETIME default CURRENT_TIMESTAMP not null,
    updated_at DATETIME default CURRENT_TIMESTAMP not null,
    unique (name, user_id)
);

create table book_collection
(
    collection_id INTEGER not null
        references collection,
    book_id       INTEGER not null
        references books,
    created_at    DATETIME default CURRENT_TIMESTAMP not null,
    primary key (collection_id, book_id)
);

create unique index user_email_uindex
    on user (email);

CREATE TABLE user_book (
   book_id    INTEGER NOT NULL,
   user_id    INTEGER NOT NULL,
   created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
   updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,

   rating     INTEGER,
   read       TIMESTAMP,

   PRIMARY KEY (book_id, user_id),
   FOREIGN KEY (book_id) REFERENCES books (id),
   FOREIGN KEY (user_id) REFERENCES user (id)
);