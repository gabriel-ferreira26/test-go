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
    └── notes/
        ├── note.go             # struct Note (o modelo de dados)
        ├── store.go            # armazenamento em memória (map + mutex)
        ├── store_test.go       # testes do armazenamento
        ├── handlers.go         # handlers HTTP (rotas da API)
        └── handlers_test.go    # testes dos handlers
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

| Método | Rota          | Descrição                     |
|--------|---------------|--------------------------------|
| GET    | `/notes`      | Lista todas as notas           |
| POST   | `/notes`      | Cria uma nova nota             |
| GET    | `/notes/{id}` | Busca uma nota pelo ID         |
| PUT    | `/notes/{id}` | Atualiza título/conteúdo       |
| DELETE | `/notes/{id}` | Remove uma nota                |

### Exemplos com curl

Criar uma nota:

```bash
curl -X POST http://localhost:8080/notes \
  -d '{"title":"Aprender Go","content":"Fazer uma API de notas"}'
```

Listar notas:

```bash
curl http://localhost:8080/notes
```

Buscar uma nota específica:

```bash
curl http://localhost:8080/notes/1
```

Atualizar uma nota:

```bash
curl -X PUT http://localhost:8080/notes/1 \
  -d '{"title":"Novo título","content":"Novo conteúdo"}'
```

Remover uma nota:

```bash
curl -X DELETE http://localhost:8080/notes/1
```

## Limitações conhecidas (de propósito)

Este projeto guarda as notas em memória — os dados somem quando o servidor
reinicia. Isso é intencional para manter o foco nos fundamentos de Go. Um
próximo passo natural para continuar aprendendo seria trocar o `Store` por
uma implementação com um banco de dados (ex: SQLite via `database/sql`),
sem precisar mexer nos handlers HTTP.
