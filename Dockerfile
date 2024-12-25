FROM golang:1.22.4

WORKDIR /app

COPY web /app/web
COPY go_final_project /app

EXPOSE 7540

ENV TODO_PORT=7540 TODO_DBFILE="scheduler.db" TODO_PASSWORD="12345"

CMD ["/app/go_final_project"]