# Cadastro e edição de perfis — MVP

Todas as rotas exigem `Authorization: Bearer JWT`. O servidor usa o ID e o papel
da conta autenticada. IDs públicos continuam numéricos; o aplicativo precisa
adaptar seus modelos Firebase ao contrato abaixo.

## Professor

`POST /teachers/` completa o primeiro cadastro do professor autenticado.
`PUT /teachers/:id` substitui os campos editáveis do próprio perfil.

```json
{
  "name": "Nome do professor",
  "photo_url": null,
  "siape": "0123456",
  "areas": [
    {"areas_id": 1, "type": "técnica"},
    {"areas_id": 2, "type": "interesse"}
  ]
}
```

Nome, SIAPE com sete dígitos e pelo menos uma área de cada tipo são obrigatórios.
Os IDs das áreas devem existir. Uma área não pode aparecer duas vezes na mesma
lista, conforme o modelo atual de associação. O aplicativo deve resolver os
nomes selecionados para IDs da API; a identidade e criação das áreas não mudam
nesta etapa.

O primeiro cadastro define `is_active=true` no servidor. Repetir POST retorna 409;
editar não altera a ativação. Conta desativada não pode se reativar pelo cadastro.
Nome, foto, SIAPE e associações são persistidos na mesma transação.

SIAPE passa de número para texto no banco e nas respostas, preservando zeros à
esquerda em novos cadastros. A migração usa o mecanismo de AutoMigrate existente.
Zeros perdidos em dados antigos não são reconstruídos automaticamente.

## Demandante

`POST /applicants/` completa o primeiro cadastro do demandante autenticado.
`PUT /applicants/:id` substitui os campos editáveis do próprio perfil.

```json
{
  "name": "Nome do demandante",
  "photo_url": null,
  "phone": "51999999999",
  "type": "pessoaFisica",
  "cpf_cnpj": "12345678900",
  "address": "Endereço completo"
}
```

Nome, telefone, tipo, documento e endereço são obrigatórios. `type` usa os valores
do Flutter: `empresa`, `orgaoPublico`, `ong`, `instituicaoEducacional`,
`pessoaFisica` ou `academico`. Dados legados com outros tipos precisam ser
regularizados pelo fluxo de edição. Neste MVP, telefone, documento e endereço
são validados quanto ao preenchimento; não se verifica titularidade nem dígitos
verificadores do documento.

Strikes começam em zero e não são alterados pela edição do perfil. Contas com
três ou mais strikes continuam bloqueadas conforme a regra de sessão existente.

## Leitura e contrato comum

- `GET /teachers/:id` e `GET /applicants/:id`: próprio usuário ou administrador.
- `GET /teachers/`, `GET /applicants/` e `/users`: acesso administrativo.
- As respostas de perfil incluem `id`, `name`, `email`, `role`, `photo_url` e
  `created_at`, além dos campos específicos. Áreas docentes incluem `areas_id`,
  `name` e `type`. O papel administrativo da API continua `admin` (no Flutter,
  `administrador`); esse mapeamento deve ser feito no cliente.
- `POST` retorna 201 com `teacherID` ou `applicantID`; `PUT` retorna 204.
- Campos desconhecidos na entrada, incluindo `id`, `role`, `email`, `strikes`,
  `is_active` e objetos GORM aninhados, retornam 400.
- `photo_url` deve ser HTTP/HTTPS quando informada. `null` ou omissão remove a
  foto personalizada no PUT. Upload de arquivos fica para etapa posterior.
- PUT recebe o formulário completo: não é atualização parcial.
- Perfil inexistente retorna 404; cadastro duplicado, 409; acesso indevido, 403.
- DELETE de perfil não é mais registrado: exclusão e gestão administrativa
  precisam de regras próprias e auditoria, ainda fora deste escopo.

## Alteração exclusiva de áreas

`PUT /teachers/:id/areas` e `PUT /areas-teachers/professor/:id` aceitam o mesmo
corpo `{"areas":[{"areas_id":1,"type":"técnica"}, ...]}` e substituem a lista
inteira. Somente o próprio professor pode escrever. A operação é transacional e
serializa alterações concorrentes do mesmo professor. O antigo corpo com apenas
`areas_id` e `type` na raiz não é mais aceito.

A lista vazia é aceita nessa operação específica e remove todas as associações.
Nesse caso, `/me` volta a indicar cadastro incompleto. As recomendações já
excluem professores sem áreas utilizáveis.

## Cadastro completo e funcionalidades futuras

`GET /me` exige nome, SIAPE válido e ambos os tipos de área para considerar o
professor completo; para demandante, exige nome e os campos descritos acima.
O estado é calculado a cada consulta. Incompleto não é sinônimo de bloqueado.

Ativação/desativação administrativa, julgamento de denúncias, reversão de
banimento e trilha de auditoria ainda serão revisados e implementados em suas
etapas. Esta entrega protege esses campos dos fluxos de edição próprios.

O teste `TestProfilePersistencePostgres` usa tabelas temporárias e requer
`TEST_DATABASE_URL`. Ele verifica ativação inicial, preservação de SIAPE e
strikes, rejeição de cadastro duplicado e rollback quando uma área não existe.
