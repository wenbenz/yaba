# Yet Another Budgeting App (YABA)
![BuildStatus](https://img.shields.io/github/actions/workflow/status/wenbenz/yaba/build.yml)
![Codecov](https://img.shields.io/codecov/c/github/wenbenz/yaba)

Tracking spending is a pain. Most budgeting tools require manual input of each
transaction, and budget labels don't always line up with credit card labels.
Financial information is also deeply personal, and with data concerns these
days, it's difficult to trust that the data will be stored with the level of
care that is warranted. This is why YABA was created. YABA intends to solve
these problems by:
- offering a budgeting tool that can be self-hosted
- creating features that enable users to define speding categories and track
spending over time
- display the credit card reward categories of actual spending to find the best
credit card for the budget
- reduce manual input by adding PFD imports and automatic assignments for
similar transactions.

## Quick Start
Build the YABA docker image
```sh
docker build --tag yaba .
```

You may consider the following flags for debugging
```sh
docker build --tag yaba --progress plain --no-cache .
```

**Change the password in `docker.env`**

Set up containers. This will bind to ports 8080 (web service) and 5432 (postgres)
```sh
docker compose up
```

Run migrations. This project uses [go-migrate](https://github.com/golang-migrate/migrate) to manage migrations.
```sh
export POSTGRES_URL='postgres://admin:password@localhost:5432/yaba?sslmode=disable'
migrate -database ${POSTGRES_URL} -path migrations up
```

## Adding database migrations
This repo uses the go-migrate framework. Migrations are run from the `server` directory.
Create a migration by running
```sh
migrate create -ext sql -dir migrations -seq MIGRATION_NAME
```

Apply a migrations by running 
```sh
migrate -database ${POSTGRES_URL} -path migrations up
```

Roll back a single migration by running
```sh
migrate -database ${POSTGRES_URL} -path server/migrations down 1
```

Note that migrations in both directions can take a number to determine how many migrations to apply to the db (as shown in the `down` command), and all of the migrations are applied if no number is specified (as in the `up` command).
