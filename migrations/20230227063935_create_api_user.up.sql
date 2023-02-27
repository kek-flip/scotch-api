CREATE USER api;
ALTER ROLE api PASSWORD 'api_password';
GRANT pg_read_all_data TO api;
GRANT pg_write_all_data TO api;