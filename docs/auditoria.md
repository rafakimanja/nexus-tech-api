# Auditoria técnica da API

**Data:** 02/11/2026  
**Escopo:** análise estática da estrutura, dos principais fluxos e tentativa de execução dos testes.  
**Situação:** registro de melhorias para avaliação e implementação futura; os itens abaixo não foram corrigidos nesta auditoria.

## Padrão identificado

A API está organizada por domínio em `internal/domain`. Cada domínio costuma dividir suas responsabilidades em arquivos de modelo, repositório, serviço, handler e rotas. Os repositórios utilizam GORM; os serviços concentram regras de negócio; os handlers tratam requisições e respostas HTTP. As dependências são montadas em `cmd/api/main.go` e os grupos de rotas registrados em `internal/server/routes.go`.

Para futuras implementações, manter essa organização, respeitando o contrato existente de cada domínio. Interfaces e injeção de dependências já são utilizadas e podem facilitar testes e a separação de integrações externas.

## Achados priorizados

### 1. Crítico — Ciclo de importação e funcionalidade de embeddings incompleta

- **Evidência:** `internal/domain/demand/demand_handler.go` importa `internal/server`, que também importa o domínio `demand`. A tentativa de teste apontou `import cycle not allowed`. Em `internal/server/embedding.go`, `CompareEmbeddings` está vazio.
- **Impacto:** o ciclo impede a compilação dos pacotes envolvidos; a comparação de embeddings ainda não entrega resultados.
- **Melhoria sugerida:** mover a integração de embeddings para um pacote próprio, definir uma interface consumida pelo domínio e completar a comparação e seu contrato de resposta. Revisar a montagem das dependências e adicionar testes para o fluxo.

### 2. Alta — Autorização por perfil e propriedade dos recursos

- **Evidência:** `internal/server/middleware.go` valida o token e coloca `userID` no contexto. Nos handlers examinados, não foram identificadas verificações correspondentes de perfil ou propriedade para as operações por ID.
- **Impacto:** dependendo da rota e das regras esperadas, um usuário autenticado pode consultar ou alterar dados de terceiros.
- **Melhoria sugerida:** definir uma matriz de permissões por perfil e operação; aplicar verificações de propriedade e autorização nos serviços ou em componentes apropriados; testar acessos permitidos e negados.

### 3. Alta — Configuração do segredo JWT

- **Evidência:** `internal/domain/user/user_token.go` procura a variável `JWT_SECRECT_KEY`, com grafia incomum, e utiliza seu valor sem verificar se está vazio.
- **Impacto:** uma configuração ausente ou digitada de forma diferente pode fazer a aplicação assinar e validar tokens com chave vazia.
- **Melhoria sugerida:** validar a presença e a força adequada do segredo na inicialização; documentar o nome da variável. Caso a grafia seja corrigida, planejar a compatibilidade com ambientes existentes e a rotação de tokens.

### 4. Alta — Migrações automáticas e remoção de colunas

- **Evidência:** `internal/server/server.go` executa `AutoMigrate` e pode remover colunas durante a inicialização; os erros dessas chamadas não são todos verificados. O comando de seed também executa migrações.
- **Impacto:** falhas de esquema podem passar despercebidas, e alterações automáticas ou destrutivas dificultam a implantação segura e a recuperação de dados.
- **Melhoria sugerida:** adotar migrações versionadas e controladas, tratar erros explicitamente, testar a evolução a partir de esquemas existentes e planejar backup e reversão para alterações destrutivas.

### 5. Alta — Consistência na criação de usuário durante o login

- **Evidência:** em `internal/server/routes.go`, o erro de `GetByEmail` não é tratado antes da decisão de criar um usuário. A criação de `User` e `UserAuth` ocorre em operações separadas.
- **Impacto:** erros de banco podem ser confundidos com ausência de usuário; uma falha parcial pode deixar a conta sem seu registro de autenticação.
- **Melhoria sugerida:** distinguir registro não encontrado de outros erros, criar conta e autenticação em uma transação e testar falhas parciais e tentativas simultâneas. Revisar também as restrições de unicidade necessárias.

### 6. Alta — Atualização de áreas do professor

- **Evidência:** `TeacherService.UpdateAreas`, em `internal/domain/teacher/teacher_service.go`, passa uma única associação para `ReplaceByProfessorID`, que remove as associações existentes antes de recriar a lista em `internal/domain/areas_teacher/areas_teacher_repository.go`.
- **Impacto:** uma operação entendida como adição ou atualização de uma área pode substituir as demais associações do professor.
- **Melhoria sugerida:** esclarecer o contrato da rota. Se a intenção for adicionar ou atualizar uma associação, implementar uma operação específica; se for substituir todas, receber e documentar a lista completa. Cobrir ambos os casos relevantes com testes.

### 7. Média — Validação, contratos HTTP e tratamento de erros

- **Evidência:** diversos handlers fazem bind diretamente em modelos persistidos e retornam `err.Error()` ao cliente. Há respostas `500` para situações que podem corresponder a entrada inválida ou registro inexistente; por exemplo, em `internal/domain/user/user_handler.go`.
- **Impacto:** campos internos podem ficar expostos à entrada, detalhes de implementação podem aparecer em respostas e clientes recebem códigos pouco precisos.
- **Melhoria sugerida:** criar DTOs de entrada e saída quando apropriado, validar campos e regras de negócio, padronizar o mapeamento de erros para `400`, `404`, `409` e `500`, e registrar detalhes internos apenas nos logs.
- **Atenção adicional:** repositórios como `internal/domain/teacher/teacher_repository.go` usam `Updates` com struct; revisar alterações de valores zero, como `false`, que podem ser ignorados pelo GORM, e definir a semântica de atualização.

### 8. Média — Cobertura de testes

- **Evidência:** foi encontrado um teste automatizado em `internal/server/routes_test.go`, voltado à rota inicial.
- **Impacto:** alterações em autenticação, autorização, regras de negócio e persistência têm pouca proteção contra regressões.
- **Melhoria sugerida:** priorizar testes de serviços e handlers para os fluxos de maior risco: autenticação, permissões, criação transacional no login, associação de áreas e tratamento de erros. Acrescentar testes de integração com banco para migrações e operações transacionais.

### 9. Média — Documentação e configuração de ambiente

- **Evidência:** o `README.md` descreve comandos básicos, mas não reúne todos os requisitos e configurações; CORS e callback OAuth possuem endereços locais definidos no código.
- **Impacto:** preparar ambientes, integrar o frontend e implantar a aplicação exige descobrir detalhes no código.
- **Melhoria sugerida:** documentar variáveis de ambiente, requisitos e preparação do PostgreSQL com pgvector, configuração do Google, execução, testes e exemplos de chamadas. Tornar origens CORS e callback configuráveis por ambiente. Fixar uma versão apropriada da imagem de banco para tornar os ambientes reproduzíveis.

## Verificação realizada e limites

Foi tentado `go test ./...`. A execução identificou o ciclo de importação acima, mas também encontrou restrições de escrita no cache de dependências do Go. Assim, a suíte não pôde ser executada integralmente e a análise não comprova o comportamento de todas as rotas em execução.

## Sequência sugerida

1. Restabelecer a compilação e definir o limite do pacote de embeddings.
2. Corrigir a configuração JWT e estabelecer autorização por perfil e recurso.
3. Proteger a integridade do login e revisar a operação de áreas de professores.
4. Introduzir migrações controladas e testes dos fluxos críticos.
5. Padronizar validações e respostas e completar a documentação operacional.
