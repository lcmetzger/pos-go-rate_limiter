# Middleware / RateLimiter

Este repositório apresenta o desenvolvimento de um `Middleware/RateLimiter` como parte do desafio proposto para cumprimento da Pós-graduação utilizando  linguagem de programação [Go](https://go.dev).

## Estrutura do projeto

O projeto está organizado com os seguintes arquivos fonte:

- [cmd/main.go](#cmdmaingo): ponto de entrada de execução da solução;
- [internal/middeware/middleware.go](#internalmiddewaremiddlewarego): implementação do Middleware que faz a interceptação das requisições;
- [internal/repository/interface_repository.go](#internalrepositoryinterface_repositorygo): interface para estabelecer o contrato de implementação dos repositórios;
- [internal/repository/pgsql_ratelimiter_repository.go](#internalrepositorypgsql_ratelimiter_repositorygo): implementação do repositório usando o banco de dados Postgres;
- [internal/repository/redis_ratelimiter_repository.go](#internalrepositoryredis_ratelimiter_repositorygo): implementação do repositório usando o banco de dados Redis;
- [rl/ratelimiter.go](#internalrepositoryredis_ratelimiter_repositorygo): implementação do RateLimiter, que faz o controle das requisições.

---

## Como configurar

A configuração do ambiente para executar a solução, se dá através de variáveis de embiente configuradas no arquivo `docker-compose.yaml`, sendo possível configurar as seguintes variáveis:

- **`REPO_TYPE`**, escolher entre as opções `PGSQL` (Postgres) ou `REDIS`;
- **`ADDR`**, deve definir a respectiva string de conexão escolhida em `REPO_TYPE`, exemplos:
  - **`PGSQL`**: _postgresql://ratelimit_user:ratelimit_pass@postgres:5432/postgres?sslmode=disable_;
  - **`REDIS`**: _redis:6379_
- **`RATE_LIMIT_TYPE`**, escolher entre `IP` ou `TOKEN`. Caso seja escolhido `TOKEN`, então obrigatóriamente deverá ser informado o parâmetro `API_KEY` no `HEADER` da requisição;
- **`IP_RATE_LIMIT`**, definir a quantidade limite de requisições aceitas no intervalo de um segundo, quando `RATE_LIMIT_TYPE` estiver definido como `IP`, exemplo: `100`
- **`TOKEN_RATE_LIMIT`**, definir a quantidade limite de requisições aceitas no intervalo de um segundo, quando `RATE_LIMIT_TYPE` estiver definido como `TOKEN`
- **`BLOCK_DURATION`**, definir o tempo em segundos, ao qual o `Middleware/Ratelimiter` deve recusar novas requisições, exemplo: `3`.

### docker-compose.yaml

Segue um exemplo de configuração. Para facilitar, basta comentar/descomentar as configurações de `REPO_TYPE`, `ADDR` e `RATE_LIMIT_TYPE`:

```yaml
version: '3.8'

services:

  postgres:
    image: postgres:latest
    environment:
      POSTGRES_USER: ratelimit_user
      POSTGRES_PASSWORD: ratelimit_pass
    ports:
      - "5454:5432"

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
  
  app:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - REPO_TYPE=PGSQL
      - ADDR=postgresql://ratelimit_user:ratelimit_pass@postgres:5432/postgres?sslmode=disable
      # - REPO_TYPE=REDIS
      # - ADDR=redis:6379
      - RATE_LIMIT_TYPE=IP
      # - RATE_LIMIT_TYPE=TOKEN
      - IP_RATE_LIMIT=100
      - TOKEN_RATE_LIMIT=100
      - BLOCK_DURATION=3
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
```

## Como executar

Primeiramente é necessário subir a aplicação usando o docke-compose com o seguinte comando:

```bash
docker-compose up --build --force-recreate
```

Executando requisições para o servidor WEB, usando a ferramenta [Apache AB](https://httpd.apache.org/docs/2.4/programs/ab.html):

```bash
ab -n 210 -B 127.0.0.1 -H "API_KEY: minha_key" http://localhost:8080/
```

Exemplo de resultado de execução do [Apache AB](https://httpd.apache.org/docs/2.4/programs/ab.html), em que foram realizadas 210 requisições, e dentre elas 110 requisições foram recusadas pelo `RateLimiter`:

```bash
This is ApacheBench, Version 2.3 <$Revision: 1913912 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 100 requests
Completed 200 requests
Finished 210 requests


Server Software:        
Server Hostname:        localhost
Server Port:            8080

Document Path:          /
Document Length:        13 bytes

Concurrency Level:      1
Time taken for tests:   0.296 seconds
Complete requests:      210
Failed requests:        110
   (Connect: 0, Receive: 0, Length: 110, Exceptions: 0)
Non-2xx responses:      110
Total transferred:      41600 bytes
HTML transferred:       11750 bytes
Requests per second:    709.54 [#/sec] (mean)
Time per request:       1.409 [ms] (mean)
Time per request:       1.409 [ms] (mean, across all concurrent requests)
Transfer rate:          137.26 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:     1    1   5.9      1      86
Waiting:        1    1   5.8      1      85
Total:          1    1   5.9      1      86

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      1
  95%      1
  98%      2
  99%      2
 100%     86 (longest request)
 ```

## Documentação das classes e funcionalidades

### `cmd/main.go`

Ponto de entrada da aplicação, aqui são obtidas todas as variáveis de ambiente definidas no arquivo `docker-compose.yaml`. Aqui também será instanciado o repositório configurado, conforme a escolha entre Postgres ou Redis (`REPO_TYPE`).

```go
  package main

  import (
    "log"
    "net/http"
    "os"
    "strconv"

    "github.com/lcmetzger/rate_limiter/internal/middleware"
    "github.com/lcmetzger/rate_limiter/internal/repository"
    "github.com/lcmetzger/rate_limiter/internal/rl"
  )

  func main() {
    // Carregar variáveis de ambiente
    addr := os.Getenv("ADDR")
    ipRateLimit, _ := strconv.ParseInt(os.Getenv("IP_RATE_LIMIT"), 10, 64)
    tokenRateLimit, _ := strconv.ParseInt(os.Getenv("TOKEN_RATE_LIMIT"), 10, 64)
    blockDuration, _ := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
    rateLimitType := os.Getenv("RATE_LIMIT_TYPE")
    repositoryType := os.Getenv("REPO_TYPE")

    var repo repository.IRateLimiterRespository = nil
    if repositoryType == "REDIS" {
      repo = repository.NewRedisRepository(addr)
    }
    if repositoryType == "PGSQL" {
      repo = repository.NewPgSqlRepository(addr)
    }
    if repo == nil {
      panic("Definir o tipo de repositorio a ser utilizado")
    }

    rateLimiter := rl.NewRateLimiter(repo, ipRateLimit, tokenRateLimit, blockDuration, rateLimitType)

    log.Printf("IpRateLimit: %v", ipRateLimit)
    log.Printf("TokenRateLimit: %v", tokenRateLimit)
    log.Printf("BlockDuration: %v", blockDuration)
    log.Printf("RateLimitType: %v", rateLimitType)
    log.Printf("RedisAddr: %v", addr)

    log.Println("Server is running...")

    mux := http.NewServeMux()
    mux.Handle("/", middleware.RateLimiterMiddleware(rateLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("Hello, world!"))
    })))

    http.ListenAndServe(":8080", mux)
  }
```

---

### `internal/middeware/middleware.go`

Implementação do `Middleware` que faz a interceptação das requisições recebidas pelo servbidor WEB, repassando o controle para o Ratelimiter que foi instanciado e injetado no request na `main.go`.
O `Middleware` invoca o `RateLimiter`, repassando os limites definidos nos parâmetors, bem como a chave escolhida entre `IP` ou `TOKEN`.

```go
package middleware

import (
  "net"
  "net/http"
  "strings"

  "github.com/lcmetzger/rate_limiter/internal/rl"
)

func RateLimiterMiddleware(rateLimiter *rl.RateLimiter, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil || ip == "" {
      ip = r.RemoteAddr
    }
    if ip == "" {
      http.Error(w, "IP address not found", http.StatusBadRequest)
      return
    }
    token := r.Header.Get("API_KEY")

    var key string
    var limit int64

    if strings.ToUpper(rateLimiter.RateLimitType) == "IP" {
      key = "ip:" + ip
      limit = rateLimiter.IpRateLimit
    }

    if strings.ToUpper(rateLimiter.RateLimitType) == "TOKEN" {
      key = "token:" + token
      limit = rateLimiter.TokenRateLimit
    }

    if rateLimiter.IsBlocked(key) || !rateLimiter.Allow(key, limit) {
      http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
      return
    }

    next.ServeHTTP(w, r)
    })
}
```

---

### `internal/repository/interface_repository.go`

Interface que define o contrato de implmentação de um repositório. Quando o repositório é implementado, ele deve implementar todos os métodos constantes no contrato. O Contrato basicamente é um CRUD.

```go
package repository

import "context"

type IRateLimiterRespository interface {
  Save(ctx context.Context, key, value string)
  Update(ctx context.Context, key, value string)
  Find(ctx context.Context, key string) (string, error)
  Delete(ctx context.Context, key string) bool
}
```

---

### `internal/repository/pgsql_ratelimiter_repository.go`

Este repositório implementa o CRUD definido na interface `IRateLimiterRespository`, usando o banco de dados Postgres.

O construtor NewPgSqlRepository cria uma nova instância do repositório PostgreSQL. Ele estabelece a conexão com o banco de dados, verifica a conectividade e cria a tabela tb_limiters se ela não existir.

```go
package repository

import (
  "context"
  "database/sql"

  _ "github.com/lib/pq"
)

const (
  sqlCreate_table = `
      CREATE TABLE IF NOT EXISTS tb_limiters (
          chkey VARCHAR(100) PRIMARY KEY,
          chvalue VARCHAR(100) NOT NULL
        );`

  sqlInsert = `
        INSERT INTO tb_limiters (chkey, chvalue)
        VALUES ($1, $2)`

  sqlUpdate = `
        UPDATE tb_limiters
        SET chvalue = $1
        WHERE chkey = $2`

  sqlSelect = `
        SELECT chvalue 
        FROM tb_limiters
        WHERE chkey = $1`

  sqlDelete = `
        DELETE 
        FROM tb_limiters
        WHERE chkey = $1`
)

type PgsqlRepository struct {
  database *sql.DB
}

func NewPgSqlRepository(addr string) *PgsqlRepository {
  db, err := sql.Open("postgres", addr)
  if err != nil {
    panic(err)
  }

  err = db.Ping()
  if err != nil {
    panic(err)
  }

  _, err = db.Exec(sqlCreate_table)
  if err != nil {
    panic(err)
  }

  return &PgsqlRepository{
    database: db,
  }
}

func (repo *PgsqlRepository) Save(ctx context.Context, key, value string) {
  _, err := repo.database.ExecContext(ctx, sqlInsert, key, value)
  if err != nil {
    panic(err)
  }
}

func (repo *PgsqlRepository) Update(ctx context.Context, key, value string) {
  _, err := repo.database.ExecContext(ctx, sqlUpdate, value, key)
  if err != nil {
    panic(err)
  }
}

func (repo *PgsqlRepository) Find(ctx context.Context, key string) (string, error) {
  var value string
  err := repo.database.QueryRowContext(ctx, sqlSelect, key).Scan(&value)
  if err != nil {
    if err == sql.ErrNoRows {
      return "", nil
    }
    panic(err)
  }
  return value, nil
}

func (repo *PgsqlRepository) Delete(ctx context.Context, key string) bool {
  _, err := repo.database.ExecContext(ctx, sqlDelete, key)
  if err != nil {
    panic(err)
  }
  return true
}
```

---

### `internal/repository/redis_ratelimiter_repository.go`

Este repositório implementa o CRUD definido na interface `IRateLimiterRespository`, usando o banco de dados Redis.

O construtor NewRedisRepository cria uma nova instância do repositório Redis, estabelece a conexão.

```go
package repository

import (
  "context"

  "github.com/redis/go-redis/v9"
)

type RedidRepository struct {
  RedisClient *redis.Client
}

func NewRedisRepository(addr string) *RedidRepository {
  return &RedidRepository{
    RedisClient: redis.NewClient(&redis.Options{
      Addr: addr,
    }),
  }
}

func (repo *RedidRepository) Save(ctx context.Context, key, value string) {
  err := repo.RedisClient.Set(ctx, key, value, 0).Err()
  if err != nil {
    panic(err)
  }
}

func (repo *RedidRepository) Update(ctx context.Context, key, value string) {
  err := repo.RedisClient.Set(ctx, key, value, 0).Err()
  if err != nil {
    panic(err)
  }
}

func (repo *RedidRepository) Find(ctx context.Context, key string) (string, error) {
  res, err := repo.RedisClient.Get(ctx, key).Result()
  if err == redis.Nil {
    return "", nil
  }
  return res, nil
}

func (repo *RedidRepository) Delete(ctx context.Context, key string) bool {
  err := repo.RedisClient.Del(ctx, key).Err()
  return err == nil
}
```

---

### `rl/ratelimiter.go`

Este é o principal mecanismo de controle das requisições da solução, que está dividida em duas funções distintas:

Este é o principal mecanismo de controle das requisições da solução, que está dividido em duas funções distintas:

**`Allow`**

A função Allow verifica se uma requisição é permitida com base no limite de taxa definido. Aqui está uma descrição detalhada do seu funcionamento:

1. Incremento do Contador
    - A função tenta encontrar o contador de requisições associado à chave (key) no repositório;
    - Se o contador não existir, ele é inicializado com o valor 1 e salvo no repositório;
    - Se o contador já existir, ele é incrementado e atualizado no repositório.

1. Configuração do Temporizador:
    - Se esta é a primeira requisição (contador igual a 1), uma goroutine é criada para aguardar o tempo de bloqueio (BlockDuration) e, em seguida, remover o contador do repositório;

1. Verificação do Limite:
    - Se o contador exceder o limite definido (limit), a função verifica se a chave está bloqueada;
    - Se a chave não estiver bloqueada, ela é marcada como bloqueada no repositório, e uma goroutine é criada para remover o bloqueio após o tempo de bloqueio (BlockDuration).
    - A função retorna false indicando que a requisição foi bloqueada.

1. Permissão da Requisição:
    - Se o contador não exceder o limite, a função retorna true, permitindo a requisição.

**`IsBlocked`**

A função IsBlocked verifica se uma chave específica (IP ou Token) está bloqueada. Aqui está uma descrição detalhada do seu funcionamento:

1. Verificação do Estado de Bloqueio:
    - A função tenta encontrar o estado de bloqueio associado à chave (key) no repositório;
    - Se a chave não estiver bloqueada, a função retorna false;
    - Se a chave estiver bloqueada, a função retorna true;

1. Tratamento de Erros:
    - Se ocorrer um erro ao tentar encontrar o estado de bloqueio, a função assume que a chave não está bloqueada e retorna false.

```go
package rl

import (
  "context"
  "log"
  "strconv"
  "strings"
  "time"

  "github.com/lcmetzger/rate_limiter/internal/repository"
)

type RateLimiter struct {
  repo           repository.IRateLimiterRespository
  IpRateLimit    int64
  TokenRateLimit int64
  BlockDuration  int
  RateLimitType  string
}

func NewRateLimiter(repository repository.IRateLimiterRespository, ipRateLimit, tokenRateLimit int64, blockDuration int, rateLimitType string) *RateLimiter {
  if strings.ToUpper(rateLimitType) != "IP" && strings.ToUpper(rateLimitType) != "TOKEN" {
    panic("O tipo de rate limiter deve ser definido através da variável de ambiente RATE_LIMIT_TYPE")
  }

  return &RateLimiter{
    repo:           repository,
    IpRateLimit:    ipRateLimit,
    TokenRateLimit: tokenRateLimit,
    BlockDuration:  blockDuration,
    RateLimitType:  rateLimitType,
  }
}

func (rl *RateLimiter) Allow(key string, limit int64) bool {
  ctx := context.Background()
  var count int64 = 1

  res, err := rl.repo.Find(ctx, key)
  if err != nil {
    return false
  }
  if res == "" {
    rl.repo.Save(ctx, key, strconv.FormatInt(count, 10))
  } else {
    count, err = strconv.ParseInt(res, 10, 64)
    if err != nil {
    panic("erro de conversão")
    }
    count++
    rl.repo.Update(ctx, key, strconv.FormatInt(count, 10))
  }

  if count == 1 {
    go func(k string) {
    time.Sleep(time.Duration(rl.BlockDuration) * time.Second)
    rl.repo.Delete(ctx, k)
    }(key)
  }

  if count > limit {
    res, err := rl.repo.Find(ctx, key+":blocked")
    if err != nil {
    log.Println(err)
    }
    if res != "true" {
    rl.repo.Save(ctx, key+":blocked", "true")
    go func(k string) {
      time.Sleep(time.Duration(rl.BlockDuration) * time.Second)
      rl.repo.Delete(ctx, k)
    }(key + ":blocked")
    }
    return false
  }
  return true
}

func (rl *RateLimiter) IsBlocked(key string) bool {
  ctx := context.Background()
  blocked, err := rl.repo.Find(ctx, key+":blocked")
  if err == nil {
    return false
  }
  return blocked == "true"
}
```
