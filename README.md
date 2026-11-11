# Nexus Tech API

A Nexus Tech API conecta pessoas e organizações que precisam de soluções a
professores que podem analisar e desenvolver essas demandas. Ela fornece ao
aplicativo uma interface HTTP para autenticação, perfis, publicação e acompanhamento
de demandas, recomendações, notificações e moderação de denúncias.

O fluxo começa com a autenticação Google. A API valida a identidade e entrega um
JWT; o aplicativo usa esse token nas chamadas seguintes. Demandantes publicam
necessidades, professores consultam e assumem demandas, e administradores
acompanham denúncias. As informações ficam no PostgreSQL. Vetores produzidos pelo
Ollama ajudam a recomendar demandas às áreas dos professores. Notificações são
armazenadas na API e podem ser enviadas via Firebase Cloud Messaging (FCM).

## Documentação

- [Arquitetura](docs/arquitetura.md): componentes, dados e principais fluxos.
- [Rotas](docs/rotas.md): autenticação, permissões e contratos principais.
- **Interface interativa:** acesse `http://localhost:3333/docs` com a API em
  execução. A especificação OpenAPI também está disponível em `/openapi.json`.
  A interface permite executar requisições reais; prefira uma base de testes.

## Executar localmente

1. Instale Go 1.26.1, PostgreSQL com a extensão `pgvector` e, para embeddings,
   Ollama com o modelo `mxbai-embed-large`.
2. Copie `.env.example` para `.env` e preencha as configurações. O exemplo não
   contém credenciais reais. `JWT_SECRET_KEY` e `SESSION_SECRET` devem ter pelo
   menos 32 caracteres. Configure um projeto OAuth Google válido.
3. Configure `DATABASE_URL` e deixe `LOCAL_DB=false` ou vazio. A API executa
   migrações ao iniciar; use primeiro um banco de desenvolvimento.
4. Defina `PORT=3333`, `OLLAMA_URL`, `GOOGLE_CLIENT_ID` e
   `GOOGLE_CLIENT_SECRET`. O login OAuth manual usa o callback fixo
   `http://localhost:3333/auth/google/callback`; cadastre-o no Google.
5. Ajuste `CORS_ALLOWED_ORIGINS` para a origem do aplicativo web, por exemplo
   `http://localhost:6170`. Para várias origens, separe por vírgulas. Sem a
   variável, a API mantém o padrão anterior `http://localhost:5173`.
6. Execute `go run ./cmd/api` e abra `/health` e `/docs`.

Alternativamente, `LOCAL_DB=true` usa `BLUEPRINT_DB_USERNAME`,
`BLUEPRINT_DB_PASSWORD`, `BLUEPRINT_DB_HOST`, `BLUEPRINT_DB_PORT`,
`BLUEPRINT_DB_DATABASE` e, opcionalmente, `BLUEPRINT_DB_SCHEMA`.
`go run ./cmd/seed` altera o esquema e cria dados de demonstração; use somente
em banco apropriado. `go run ./cmd/embeddings` reprocessa vetores pendentes.

FCM é opcional: configure `FCM_PROJECT_ID` e credenciais Google válidas por
`GOOGLE_APPLICATION_CREDENTIALS` para ativar o worker de push. A central HTTP
de notificações funciona sem FCM. Mantenha o JSON da conta de serviço fora do
Git e da imagem; em contêiner, monte-o como segredo.

## Testes

Execute `go test ./...`. Testes PostgreSQL são pulados sem `TEST_DATABASE_URL`;
para executá-los, configure essa variável para um banco de testes onde a conta
possa criar e remover schemas. A aplicação usa `AutoMigrate` e não possui
rollback automático de migrações.
