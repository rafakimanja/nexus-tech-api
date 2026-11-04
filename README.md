# Nexus Tech API

API em Go para autenticação, perfis, áreas e demandas do Nexus Tech.

**[Documentação completa da API](docs/README.md)**: configuração, arquitetura,
fluxos, referência de rotas, contratos e testes.

## Getting Started

O contrato de login, sessão de 12 horas e consulta `GET /me` está em
[docs/autenticacao.md](docs/autenticacao.md).

Cadastro, edição e permissões de perfis estão em [docs/perfis.md](docs/perfis.md).

Contratos e ciclo de vida de demandas estão em [docs/demandas.md](docs/demandas.md).

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```

Populate the database independently from the API (requires the same database
environment variables as the API, loaded from `.env` or the shell):
```bash
make seed
```
This creates/updates the schema and inserts sample areas, applicants, teachers
and demands. It is safe to run more than once. To run directly:
```bash
go run ./cmd/seed
```

### Recomendações para professores

Configure `OLLAMA_URL` com o endereço do Ollama e disponibilize o modelo
`mxbai-embed-large` (vetores de 1024 dimensões). O PostgreSQL precisa ter a
extensão pgvector instalada e habilitada antes das migrações.

Ao criar ou alterar uma demanda ou área, a API tenta gerar seu embedding.
Se o Ollama falhar, o cadastro é mantido e o embedding fica pendente. Para
processar embeddings pendentes de áreas e demandas novamente, execute:

```bash
go run ./cmd/embeddings
```

O próprio professor autenticado pode consultar
`GET /teachers/:id/recommendations?limit=10` (limite de 1 a 100). São
consideradas demandas com status `cadastrada`, não canceladas, sem professor atribuído
e com embedding disponível. Áreas técnicas recebem peso 0,8 e áreas de
interesse, 0,5; a melhor correspondência determina a ordem. Similaridades
negativas são tratadas como zero. Professor sem áreas utilizáveis recebe uma
lista vazia. O processamento pendente precisa ser executado novamente quando
houver falhas; não há fila automática neste MVP.

Para executar o teste de integração da recomendação, configure
`TEST_DATABASE_URL` para um PostgreSQL de teste com pgvector habilitado e rode
`go test ./internal/domain/recommendation`.
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

Testes de integração requerem `TEST_DATABASE_URL` apontando para um PostgreSQL
exclusivo de testes, com pgvector habilitado. Sem essa variável, são ignorados:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
