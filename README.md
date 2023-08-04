# flood-social-rep
Silly (:3) bot for accounting chat members "social credits" (actually integers). Was made for local flood group, j4f.
If may not match your usecase because of specific method of getting reactions (recieving it on POST /reactions endpoint, in form of Telegram message JSON).

## Build and run
First, copy `.env.example` to `.env` and edit values
Then, you can build it with Docker or manually

### Docker and docker-compose
```bash
$ docker compose up -d
```

### Manually
```bash
$ go build
$ ./flood-social-rep
```
