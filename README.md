# bitburst


docker run -d -p 5432:5432 --name bitburst-postgres -e POSTGRES_PASSWORD=secret  postgres

docker exec -it bitburst-postgres psql -U postgres

CREATE DATABASE bitburst;
