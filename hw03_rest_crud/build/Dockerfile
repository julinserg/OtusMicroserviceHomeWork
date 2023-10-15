# Собираем в гошке
FROM golang:1.17.2 as build

ENV BIN_FILE /opt/users_service/users_service-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/users_service/*

# На выходе тонкий образ
FROM golang:1.17.2


LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="users_service"
LABEL MAINTAINERS="student@otus.ru"

ENV BIN_FILE "/opt/users_service/users_service-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/users_service/config.toml
COPY ./configs/config-docker.toml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
