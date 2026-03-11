<font size= "5"> **Table Of Contents** </font>
- [Todo list](#todo-list)
- [Introduction](#introduction)
- [Environment for development](#environment-for-development)
  - [Database](#database)
    - [Mysql](#mysql)
    - [Migration](#migration)
  - [Redis](#redis)
  - [Service](#service)
    - [Preparation](#preparation)
    - [Run server](#run-server)


# Todo list
- [x] Integrate FX framework
- [x] Add middleware layers to http server
- [x] Implement `users` methods use middleware in gin framework
- [x] Complete `users` CRUD methods
- [ ] Deploy the the system to a real domain (fiagram.io.vn)
- [x] Integrate mock framework for automation testing
- [x] Fixbug SQL injection when querying database directly
- [x] Apply Docker for production releases
- [ ] (Tech debt) Fork and rewrite openapi to add `RegisterWith<middleware_names>Mids` intent to group middlewares for convenience

# Introduction
- The backend service belongs to Fiagram project.

# Environment for development
## Database
### Mysql
- Start MySQL database by docker with the default configuration from `configs/local.yaml`
```
./scripts/run-docker-mysql-dev.sh
```
### Migration
- Add a new schema for database
```
make migrate-new <name_of_new_schema>
```

## Redis
- Start Redis database by docker
```
./scripts/run-docker-redis-dev.sh
```

## Service
### Preparation
- Download tools for the service
```
make init
```

### Run server
- Run service with a make command
```
make run-server
```
