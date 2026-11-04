# Pendências mínimas da API para integração com o aplicativo

Atualizado em 04/11/2026. Esta lista considera um uso básico: autenticar com
Google, completar o perfil, consultar áreas e permitir que um demandante
publique e acompanhe uma demanda e que um professor a assuma e conclua.
Os itens de demandas foram confrontados com o modelo `demanda.dart` do app.
Considera-se que o aplicativo será adaptado para IDs numéricos; esta análise
não confirma que ele já consome os contratos da API.

Login, sessão, perfis, áreas e CRUD de demandas já possuem rotas na API. Os
itens abaixo registram a implementação e o trabalho restante na API para esse
fluxo, incluindo sua configuração e validação.

## 1. Demandas — implementado

Contratos e decisões estão em [demandas.md](demandas.md). Área opcional, com
IDs existentes preservados; edição sem área não desfaz o vínculo. Testes
unitários e HTTP executados. A validação com PostgreSQL real permanece
pendente: o teste foi criado, mas exige `TEST_DATABASE_URL` configurada.

- [x] Acrescentar `publicoAlvo` e `impacto` ao modelo e aos contratos de
  criação, edição e leitura, com preenchimento obrigatório nas entradas.
- [x] Tornar `AreaID` opcional, validar a área quando informada e preservar
  os vínculos existentes quando o aplicativo editar sem enviar esse campo.
- [x] Criar contratos de entrada e saída para demandas, validar os dados e
  identificar o demandante pelo usuário autenticado, sem aceitar do cliente
  campos internos ou que ele não possa alterar. Retornar título, descrição,
  dados do demandante (ID, nome e tipo) e do professor (ID e nome, quando
  atribuído), estado e datas. Os nomes estão associados aos usuários dos
  perfis e precisam ser carregados ou consultados para compor a resposta.
- [x] Aplicar autorização por papel e propriedade nas consultas e alterações:
  o demandante gerencia suas demandas e o professor acessa a prateleira e as
  demandas atribuídas a ele. A prateleira do modelo do app inclui demandas
  `cadastrada` e `emProducao`.
- [x] Validar os estados `cadastrada`, `emAnalise`, `emProducao`, `concluida`
  e `cancelada`. Implementar as transições do app: `cadastrada` para
  `emAnalise` ou `cancelada`; `emAnalise` para `emProducao` ou `cadastrada`;
  `emProducao` para `concluida`. Os dois últimos estados são terminais.
- [x] Implementar marcação para análise, devolução, atribuição e conclusão
  pelo professor autorizado. Registrar no servidor início e prazo da análise,
  início da produção, conclusão e descrição da solução. Garantir transições
  atômicas para impedir que dois professores reservem a mesma demanda.
- [x] Fazer a análise expirar após 24 horas e devolver a demanda à prateleira,
  liberando a atribuição de forma segura mesmo após indisponibilidade da API.
- [x] Permitir edição e cancelamento pelo demandante apenas em `cadastrada`;
  preservar motivo, data e histórico do cancelamento. Conciliar `Status` e
  `IsCancelled` para evitar estados contraditórios, sem tratar exclusão lógica
  do GORM como equivalente ao cancelamento.
- [x] Disponibilizar consultas para o demandante acompanhar suas demandas e
  para o professor listar a prateleira e as atribuídas a ele.
- [x] Retornar erros adequados para dados inválidos, recurso inexistente,
  acesso negado e conflitos, sem expor detalhes internos.

## 2. Configuração e contratos

- [ ] Preparar uma configuração reproduzível do PostgreSQL com pgvector para
  iniciar a API; a imagem atual do Docker usa `postgres:latest`, enquanto o
  esquema exige a extensão. Documentar as variáveis e os passos necessários,
  sem incluir segredos.
- [ ] Documentar os contratos usados no fluxo básico, incluindo campos, IDs,
  estados e erros de demandas, em complemento à documentação existente de
  autenticação e perfis.
- [ ] Configurar as origens CORS por ambiente se o aplicativo for executado
  na Web; a configuração atual aceita apenas `http://localhost:5173`.

## 3. Validação

- [ ] Executar os testes de integração existentes de identidade e perfis com
  PostgreSQL de teste e validar o login Google real com a configuração usada
  pelo aplicativo.
- [ ] Testar o fluxo básico de ponta a ponta na API: login, preenchimento de
  perfil, consulta de áreas, criação e consulta de demanda, análise, devolução
  ou atribuição e conclusão, edição e cancelamento quando permitidos.
- [ ] Testar acessos negados, dados inválidos e atribuições simultâneas da
  mesma demanda, transições e expiração da análise, além de confirmar que uma
  falha na geração de embeddings não impede o cadastro ou a edição da demanda.

Chat, notificações, anexos, denúncias, administração avançada, recomendações,
republicação e migração de dados não fazem parte deste recorte básico. Se a
primeira entrega exigir algum desses fluxos ou preservar contas existentes,
suas dependências na API precisarão ser avaliadas separadamente.
