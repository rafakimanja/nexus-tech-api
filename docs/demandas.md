# Demandas — contrato do fluxo básico

Todas as rotas usam `Authorization: Bearer JWT`. Demandante e professor
precisam ter o respectivo perfil cadastrado. Contas bloqueadas e professores
inativos não têm acesso. IDs são numéricos; os nomes de campos abaixo são o
contrato HTTP, independentemente dos nomes internos do GORM.

## Criação e edição

`POST /demands/` cria uma demanda e retorna `201` com o DTO completo.
`PUT /demands/:id` edita o conteúdo e retorna `200` com o mesmo DTO.
Ambos recebem:

```json
{
  "titulo": "Agendamento de salas",
  "descricao": "Permitir consulta e reserva de salas disponíveis.",
  "publicoAlvo": "Comunidade acadêmica",
  "impacto": "Reduzir conflitos de reservas",
  "areaId": null
}
```

Os quatro textos são obrigatórios e não podem estar vazios. `areaId` é opcional;
quando informado, deve ser um ID positivo de uma área existente. Na criação,
ausência ou `null` significa demanda sem área. Na edição, ausência ou `null`
preserva a área existente, permitindo que o app, que não possui esse campo,
edite demandas antigas sem desfazer vínculos. Um ID informado substitui a área.

Somente o demandante pode criar demandas e editar as próprias enquanto estiverem
`cadastrada`. A identidade vem da sessão. Estado inicial, atribuição e datas são
definidos no servidor. Campos desconhecidos ou internos são rejeitados com `400`;
o app deve enviar apenas os campos de entrada, e não todo o seu `toMap()`.
O corpo tem limite de 1 MiB. A indisponibilidade do serviço de embeddings não
desfaz a criação ou a edição.

## Consultas e resposta

- `GET /demands/`: demandas próprias para o demandante; prateleira para o
  professor; todas para o administrador (somente leitura).
- `GET /demands/?escopo=minhas`: demandas próprias do demandante ou atribuídas
  ao professor, incluindo as concluídas.
- `GET /demands/?escopo=prateleira`: exclusivo de professor, com estados
  `cadastrada` e `emProducao`. Estar na prateleira não significa estar disponível
  para análise: apenas `cadastrada` pode ser reservada.
- `GET /demands/:id`: o demandante consulta apenas as próprias; o professor,
  as da prateleira ou atribuídas a ele; administrador pode consultar todas.

Listas retornam um array (inclusive `[]`), ordenado por criação decrescente e
ID decrescente. O DTO de cada demanda é:

```json
{
  "id": 42,
  "demandanteId": 1,
  "demandanteNome": "Maria",
  "demandanteTipo": "pessoaFisica",
  "titulo": "Agendamento de salas",
  "descricao": "Permitir consulta e reserva de salas disponíveis.",
  "publicoAlvo": "Comunidade acadêmica",
  "impacto": "Reduzir conflitos de reservas",
  "areaId": null,
  "status": "cadastrada",
  "motivoCancelamento": null,
  "professorId": null,
  "professorNome": null,
  "analiseIniciadaEm": null,
  "analiseExpiraEm": null,
  "producaoIniciadaEm": null,
  "concluidaEm": null,
  "descricaoSolucao": null,
  "criadoEm": "2026-11-04T12:00:00Z",
  "atualizadoEm": "2026-11-04T12:00:00Z",
  "canceladoEm": null
}
```

Datas usam RFC 3339/ISO 8601. Nomes e tipo vêm dos perfis atuais; a resposta
não inclui CPF/CNPJ, telefone, e-mail, embeddings nem estruturas internas.
`demandanteUid` e `professorUid` do modelo antigo do app passam a corresponder
a `demandanteId` e `professorId`. Republicação não integra este contrato.

## Ações e estados

Todas as ações são `POST /demands/:id/<ação>`, retornando `200` com o DTO.

| Ação | Quem pode executar | Transição | Corpo |
| --- | --- | --- | --- |
| `analisar` | Professor | `cadastrada` → `emAnalise` | Vazio ou `{}` |
| `devolver` | Professor atribuído | `emAnalise` → `cadastrada` | Vazio ou `{}` |
| `assumir` | Professor atribuído | `emAnalise` → `emProducao` | Vazio ou `{}` |
| `concluir` | Professor atribuído | `emProducao` → `concluida` | `{"descricaoSolucao":"Solução entregue"}` |
| `cancelar` | Demandante proprietário | `cadastrada` → `cancelada` | `{"motivoCancelamento":"Não preciso mais"}` |

Solução e motivo são obrigatórios e não podem estar vazios. `DELETE /demands/:id`
é um alias de cancelamento e exige o mesmo corpo com motivo; nunca exclui o
registro. Estados `concluida` e `cancelada` são terminais. Cancelamento preserva
conteúdo, motivo, datas e acesso do proprietário ao registro. `IsCancelled` é
mantido internamente em sincronia com o estado, sem controle pelo cliente.
O administrador não recebe permissão para executar essas mutações.

As mutações bloqueiam a linha da demanda em uma transação. Uma reserva
concorrente só pode vencer uma vez; as demais recebem `409`. Não é permitido
assumir diretamente uma demanda cadastrada nem alterar o estado pelo `PUT`.

A reserva dura 24 horas a partir da marcação para análise. A API verifica as
reservas vencidas na inicialização, a cada minuto e antes das operações de
demandas. O processo é idempotente e recupera prazos vencidos durante paradas.
A devolução manual ou por expiração libera o professor e limpa os dois campos
de análise. A atribuição para produção revalida o prazo após adquirir o lock;
uma reserva vencida não pode ser assumida. O prazo é preservado como informação
histórica na produção, mas não causa expiração fora de `emAnalise`.

## Erros, esquema e validação

Formato: `{"error":{"code":400,"message":"..."}}`.

- `400`: entrada, ID, área ou escopo inválido; campos não permitidos.
- `401`: sessão ausente, inválida ou expirada (middleware de autenticação).
- `403`: papel, propriedade ou perfil sem permissão; conta bloqueada.
- `404`: demanda inexistente.
- `409`: transição incompatível com o estado, reserva concorrente ou vencida.
- `500`: falha interna; detalhes ficam apenas nos logs.

A inicialização acrescenta os campos ao esquema e torna `area_id` anulável,
mantendo a chave estrangeira e os vínculos existentes. Demandas anteriores
recebem textos vazios para os novos campos de conteúdo e datas opcionais nulas;
novas criações e edições exigem os textos completos. Não há conversão automática
de estados legados, importação de dados nem reconstrução de datas históricas.
O seed passa a criar novos exemplos com estados do contrato; registros já
existentes não são sobrescritos.

Testes unitários e HTTP: `go test ./internal/domain/demand`.
O teste `TestDemandPersistencePostgres` requer `TEST_DATABASE_URL` apontando
para um banco exclusivo de testes, com permissão para criar schemas. Ele cobre
persistência, isolamento das listas, rollback, cancelamento, expiração e reserva
concorrente com conexões distintas. Sem essa variável, ele é ignorado. Esse teste
isola a lógica relacional e não valida pgvector nem a migração do esquema real.
