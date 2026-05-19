ALTER TABLE stopwords DROP CONSTRAINT stopwords_user_id_fkey;
ALTER TABLE stopwords ADD CONSTRAINT stopwords_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE tags DROP CONSTRAINT tags_user_id_fkey;
ALTER TABLE tags ADD CONSTRAINT tags_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE files DROP CONSTRAINT files_user_id_fkey;
ALTER TABLE files ADD CONSTRAINT files_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE items DROP CONSTRAINT items_user_id_fkey;
ALTER TABLE items ADD CONSTRAINT items_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
