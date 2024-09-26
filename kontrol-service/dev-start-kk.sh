DB_HOSTNAME=localhost DB_USERNAME=postgres DB_NAME=kardinal DB_PORT=5432 DB_PASSWORD=postgres reflex -s -r '\.go$' -- go run main.go "$@"
