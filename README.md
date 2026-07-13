# Notes API

Uma API REST simples para um bloco de notas, escrita em Go usando apenas a
biblioteca padrão (`net/http`) — sem frameworks. O objetivo deste projeto é
servir como introdução prática à linguagem.

## Estrutura do projeto

```
.
├── go.mod                      # define o módulo e a versão do Go
├── main.go                     # ponto de entrada: cria o servidor HTTP
└── internal/
    ├── httpx/
    │   └── json.go              # helpers de resposta JSON compartilhados
    ├── auth/
    │   ├── user.go              # struct User
    │   ├── store.go             # usuários + sessões em memória (bcrypt)
    │   ├── store_test.go        # testes do armazenamento de auth
    │   ├── middleware.go        # RequireAuth: exige Bearer token válido
    │   ├── handlers.go          # rotas /auth/register e /auth/login
    │   └── handlers_test.go     # testes dos handlers de auth
    └── notes/
        ├── note.go              # struct Note (o modelo de dados)
        ├── store.go             # armazenamento em memória (map + mutex)
        ├── store_test.go        # testes do armazenamento
        ├── handlers.go          # handlers HTTP (rotas da API)
        └── handlers_test.go     # testes dos handlers
```

`internal/` é uma convenção do Go: qualquer pacote dentro dele só pode ser
importado por código do mesmo módulo. É a forma idiomática de marcar código
como "detalhe de implementação", não uma API pública.

## Conceitos de Go usados aqui

- **Structs e tags de JSON** (`note.go`): `Note` é uma struct comum; as tags
  `` `json:"..."` `` controlam como os campos são serializados.
- **Ponteiros e métodos** (`store.go`): `Store` é manipulado sempre via
  ponteiro (`*Store`), então cada método pode alterar o mapa interno.
- **Concorrência segura com `sync.RWMutex`**: o servidor HTTP do Go atende
  cada requisição em sua própria goroutine, então o mapa em memória precisa
  de um mutex para evitar leituras/escritas concorrentes inconsistentes.
- **Tratamento de erros idiomático**: funções retornam `(valor, error)`; o
  pacote define um erro sentinela (`ErrNotFound`) verificado com
  `errors.Is`.
- **`net/http` puro**: desde o Go 1.22, `http.ServeMux` já suporta métodos
  HTTP e wildcards de path (`"GET /notes/{id}"`), então não precisamos de
  um router de terceiros (como gin ou chi) para uma API simples.
- **Testes com o pacote `testing`**: `store_test.go` testa a lógica pura;
  `handlers_test.go` usa `httptest.NewServer` para testar a API por cima de
  HTTP de verdade.
- **Autenticação com `context.Context`** (`auth/middleware.go`): o
  middleware `RequireAuth` valida o header `Authorization: Bearer <token>`
  e injeta o ID do usuário autenticado no `context.Context` da requisição;
  os handlers de notas o recuperam via `auth.UserIDFromContext`. É assim
  que Go passa dados "por fora" da assinatura das funções, entre camadas.
- **Hash de senha com `golang.org/x/crypto/bcrypt`**: senhas nunca são
  guardadas em texto puro. `bcrypt` é a extensão oficial do time do Go para
  isso — nunca implemente hashing de senha na mão.
- **Tokens de sessão opacos**: em vez de JWT, o login gera um token
  aleatório (`crypto/rand`) guardado num mapa `token -> userID` no
  servidor. É mais simples de entender que JWT e evita puxar uma
  dependência extra só para assinar tokens — a troca por JWT no futuro
  ficaria isolada dentro do pacote `auth`.
- **Composição de `http.Handler`**: `RequireAuth` recebe um
  `http.Handler` e devolve outro — o padrão de middleware mais comum em
  Go, sem precisar de nenhum framework.

## Rodando o projeto

```bash
go run .
```

O servidor sobe em `http://localhost:8080`.

## Rodando os testes

```bash
go test ./...
```

## Endpoints

| Método | Rota             | Auth? | Descrição                          |
|--------|------------------|-------|--------------------------------------|
| POST   | `/auth/register` | não   | Cria uma conta (username + password) |
| POST   | `/auth/login`    | não   | Autentica e retorna um token         |
| GET    | `/notes`         | sim   | Lista as notas do usuário logado     |
| POST   | `/notes`         | sim   | Cria uma nova nota                   |
| GET    | `/notes/{id}`    | sim   | Busca uma nota pelo ID               |
| PUT    | `/notes/{id}`    | sim   | Atualiza título/conteúdo             |
| DELETE | `/notes/{id}`    | sim   | Remove uma nota                      |

Todas as rotas `/notes/*` exigem o header `Authorization: Bearer <token>`.
Cada nota pertence a quem a criou — um usuário nunca vê ou edita as notas de
outro (tentar buscar uma nota de outra pessoa retorna `404`, não `403`, para
não revelar que ela existe).

### Exemplos com curl

Criar uma conta:

```bash
curl -X POST http://localhost:8080/auth/register \
  -d '{"username":"alice","password":"s3cret"}'
```

Fazer login (retorna `{"token": "..."}`):

```bash
curl -X POST http://localhost:8080/auth/login \
  -d '{"username":"alice","password":"s3cret"}'
```

Guarde o token e envie-o em todas as chamadas a `/notes`:

```bash
TOKEN="cole-o-token-aqui"
```

Criar uma nota:

```bash
curl -X POST http://localhost:8080/notes \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Aprender Go","content":"Fazer uma API de notas"}'
```

Listar notas:

```bash
curl http://localhost:8080/notes -H "Authorization: Bearer $TOKEN"
```

Buscar uma nota específica:

```bash
curl http://localhost:8080/notes/1 -H "Authorization: Bearer $TOKEN"
```

Atualizar uma nota:

```bash
curl -X PUT http://localhost:8080/notes/1 \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Novo título","content":"Novo conteúdo"}'
```

Remover uma nota:

```bash
curl -X DELETE http://localhost:8080/notes/1 -H "Authorization: Bearer $TOKEN"
```

## Limitações conhecidas (de propósito)

Este projeto guarda notas, usuários e sessões em memória — tudo some quando
o servidor reinicia (incluindo os tokens de login). Isso é intencional para
manter o foco nos fundamentos de Go. Próximos passos naturais para continuar
aprendendo:

- Trocar o `notes.Store` por uma implementação com banco de dados (ex:
  SQLite via `database/sql`), sem precisar mexer nos handlers HTTP.
- Trocar os tokens de sessão opacos por JWT, se quiser tokens que carreguem
  sua própria validade sem consulta ao servidor.
- Adicionar expiração de sessão (hoje um token de login nunca expira).
