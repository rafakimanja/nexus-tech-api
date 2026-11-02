# Pendências para integração da API com o aplicativo

Atualizado em 02/11/2026. Base: `GUIA_AGENTES_MIGRACAO.md`, auditoria anterior e
conferência do código atual da API e das dependências Firebase do aplicativo.
Esta lista contém somente trabalho restante. Cada etapa deve ser revisada com
o responsável pelo produto antes da implementação. A ordem abaixo é sugerida;
não representa aprovação de todo o escopo para o MVP.

## 1. Demandas e ciclo de vida — próxima etapa sugerida

- [ ] Comparar campo a campo com o Flutter e acrescentar público-alvo, impacto, datas de análise/produção/conclusão, descrição da solução e vínculo de republicação.
- [ ] Definir se a área explícita da demanda será obrigatória: a API exige `AreaID`, mas o modelo/formulário do aplicativo precisa ser reconciliado com esse requisito.
- [ ] Padronizar o CRUD e os dados legados com os estados `cadastrada`, `emAnalise`, `emProducao`, `concluida` e `cancelada`.
- [ ] Criar DTOs de entrada e saída de demandas, com campos editáveis limitados, identificação do demandante pelo usuário autenticado e datas críticas geradas no servidor.
- [ ] Aplicar autorização por papel, propriedade e estado nas consultas, alterações e ações das demandas.
- [ ] Implementar ações de marcar para análise, devolver, assumir e concluir, com transições válidas e timestamps.
- [ ] Restringir edição e cancelamento pelo demandante ao estado `cadastrada`; preservar motivo e histórico no cancelamento.
- [ ] Implementar republicação como nova demanda, mantendo vínculo histórico e impedindo republicação duplicada.
- [ ] Proteger ações concorrentes com transações e locks ou atualizações condicionais.
- [ ] Confirmar e implementar expiração de análise em 24 horas por job idempotente no servidor, incluindo recuperação após indisponibilidade.
- [ ] Implementar listas por papel, filtros acumulativos, pesquisa, ordenação e paginação.
- [ ] Definir os efeitos de cada transição sobre chats, notificações e auditoria; implementar esses efeitos junto aos respectivos módulos.

## 2. Áreas, contratos e permissões restantes

- [ ] Normalizar nomes de áreas e estabelecer unicidade e comportamento para criação concorrente de áreas equivalentes.
- [ ] Definir e aplicar permissões para criar, renomear e excluir áreas, considerando os vínculos existentes.
- [ ] Revisar entradas ainda vinculadas diretamente aos modelos GORM, especialmente demandas, áreas e criação administrativa de usuários.
- [ ] Padronizar respostas de erro restantes: evitar detalhes internos e distinguir entrada inválida, ausência, conflito e falha interna.
- [ ] Consolidar OpenAPI versionado com exemplos e os contratos de IDs, enums, datas, paginação e erros.
- [ ] Completar a matriz de permissões dos recursos ainda não protegidos e dos novos módulos.
- [ ] Substituir a atribuição de administrador por e-mail fixo no código por um processo/configuração controlado; revisar elegibilidade institucional de professores.
- [ ] Definir o procedimento de vinculação controlada para contas legadas que recebem 409 no login.
- [ ] Revisar necessidade de validações adicionais de telefone e CPF/CNPJ, hoje limitadas ao preenchimento no cadastro da API.
- [ ] Remover ou adaptar handlers e serviços antigos de mutação de perfis que ficaram sem rota, para evitar reintroduzir contratos inseguros no futuro.

## 3. Administração, denúncias e auditoria

- [ ] Criar visão geral administrativa com indicadores e critérios de cálculo acordados com o aplicativo.
- [ ] Implementar registro de denúncias, proteção contra duplicação e acesso restrito aos envolvidos/autorizados.
- [ ] Implementar fila de denúncias, contador, filtros e detalhes.
- [ ] Implementar julgamento e parecer administrativo, impedindo julgamento duplo.
- [ ] Aplicar advertências e suspensão decorrentes do julgamento de forma transacional.
- [ ] Modelar o estado administrativo de banimento, data e motivo, preservando a regra do aplicativo de reversão somente pelo administrador.
- [ ] Implementar reversão administrativa de banimento e gestão de strikes com justificativa.
- [ ] Implementar ativação/desativação de professor pelo administrador, com data e motivo.
- [ ] Implementar métricas por professor conforme as definições do aplicativo.
- [ ] Completar gestão administrativa de demandantes além das consultas existentes.
- [ ] Registrar trilha de auditoria com ator, ação, alvo, data e justificativa para operações administrativas e eventos relevantes.

## 4. Chat e notificações

- [ ] Criar conversas vinculadas às demandas, participantes e mensagens.
- [ ] Implementar autorização por participante, paginação/ordenação de mensagens, contadores e marcação de leitura.
- [ ] Implementar encerramento de conversa e mensagens de sistema conforme as transições da demanda.
- [ ] Tratar envios duplicados e concorrência.
- [ ] Persistir notificações por destinatário, geradas por eventos do servidor.
- [ ] Implementar listagem, contador de não lidas e leitura individual/em lote.
- [ ] Garantir consistência entre eventos de domínio e seus efeitos; definir retentativa/outbox quando houver efeitos externos.
- [ ] Definir atualização por polling, SSE ou WebSocket, incluindo reconexão e recuperação de eventos.

## 5. Fotos e anexos

- [ ] Escolher armazenamento de objetos e política de persistência, retenção e recuperação.
- [ ] Implementar upload e metadados de fotos e anexos vinculados ao usuário ou à demanda.
- [ ] Confirmar limites/tipos do aplicativo e validar tamanho e conteúdo no servidor.
- [ ] Implementar download autorizado e definir uso de URLs temporárias ou proxy autenticado.
- [ ] Tratar falhas entre banco e armazenamento, exclusão e limpeza de objetos órfãos.

## 6. Integração no Flutter

- [ ] Criar cliente HTTP por ambiente, tratamento de erros e armazenamento seguro do JWT.
- [ ] Adaptar Google Sign-In para enviar o ID token à API com audiência correta por plataforma.
- [ ] Substituir o estado FirebaseAuth pela sessão da API, consumindo `/me`, expiração de 12 horas, nova autenticação e logout local.
- [ ] Adaptar telas de cadastro incompleto, bloqueio e conflito de vinculação de conta.
- [ ] Adaptar modelos e parsers aos DTOs, incluindo IDs numéricos, `admin`/`administrador`, campos de perfil e SIAPE textual.
- [ ] Integrar cadastro/edição de perfis e mapear áreas selecionadas para os IDs da API.
- [ ] Consumir recomendações na tela do professor e integrar sua ordenação/destaque à experiência do Kanban.
- [ ] Substituir o repositório Firebase de demandas e integrar formulários, filtros, detalhes e ações aos contratos aprovados.
- [ ] Substituir repositórios de chat, notificações, denúncias, administração, fotos e anexos conforme os módulos da API forem entregues.
- [ ] Adaptar atualização das telas, reconexão, conflitos e comportamento offline acordado.
- [ ] Remover dependências e inicialização Firebase somente após migrar e validar os fluxos que as utilizam.

## 7. Banco, configuração e operação

- [ ] Substituir migrações automáticas/destrutivas na inicialização por migrações versionadas, com tratamento de falhas e procedimento de recuperação.
- [ ] Validar evolução do esquema existente, incluindo SIAPE textual, vínculos únicos Google e professor opcional nas demandas.
- [ ] Ajustar o ambiente Docker para uma versão fixa de PostgreSQL compatível com pgvector e preparar a extensão de forma reproduzível.
- [ ] Tornar origens CORS e callback Google configuráveis por ambiente.
- [ ] Completar documentação de configuração e execução por ambiente sem expor segredos.
- [ ] Preparar TLS, logs operacionais, monitoramento, backups e ensaio de restauração.
- [ ] Definir e testar execução operacional dos jobs e do reprocessamento manual de embeddings pendentes.

## 8. Migração dos dados existentes — condicionada à necessidade de preservação

- [ ] Confirmar quais dados/contas Firebase precisam ser preservados e os volumes envolvidos.
- [ ] Inventariar coleções, subcoleções, identidades e objetos, incluindo configurações externas ou Cloud Functions ainda não mapeadas.
- [ ] Definir mapeamento de Firebase UID/document ID/Google subject para IDs PostgreSQL, com tratamento de contas ambíguas e referências quebradas.
- [ ] Implementar importação idempotente de usuários, perfis, áreas, demandas, histórico, chats, mensagens, notificações, denúncias e auditoria.
- [ ] Migrar objetos e metadados de arquivos com verificação de integridade e permissões.
- [ ] Regularizar enums legados, duplicidades de identidade e dados de perfil incompatíveis com o contrato atual.
- [ ] Ensaiar migração isolada, reconciliar contagens e relacionamentos e definir corte, janela de indisponibilidade e rollback.
- [ ] Validar dados e fluxos antes de desligar Firebase.

## 9. Validação ainda pendente

- [ ] Configurar PostgreSQL de teste e executar os testes de integração existentes de recomendações, identidades e perfis; eles ainda não foram executados neste ambiente.
- [ ] Testar migrações a partir de um esquema anterior, sem depender apenas de banco vazio.
- [ ] Cobrir concorrência real de login, substituição de áreas, ações de demanda e julgamento administrativo.
- [ ] Validar login Google real e sessão por plataforma suportada.
- [ ] Validar geração e regeneração de embeddings com Ollama real, falhas e reprocessamento.
- [ ] Avaliar a qualidade do ranking com demandas e áreas representativas; ajustar pesos somente se os resultados justificarem.
- [ ] Executar `flutter analyze`, `flutter test` e testes de contrato dos repositórios HTTP após a adaptação do app.
- [ ] Executar cenários completos de cadastro, demanda, cancelamento/republicação, chat, arquivos, denúncias e administração conforme os módulos forem migrados.
- [ ] Testar isolamento entre usuários, contas bloqueadas, falhas de rede, reconexão e recuperação de dados.

## 10. Decisões e melhorias futuras — fora do MVP imediato

- [ ] Definir quais funcionalidades restantes são obrigatórias para a primeira entrega e quais podem ser adiadas.
- [ ] Confirmar plataformas suportadas e exigências de funcionamento offline/tempo real.
- [ ] Definir infraestrutura e retenção de arquivos e dados históricos.
- [ ] Planejar notificações de novas demandas compatíveis com o perfil e regras de prioridade para professores.
- [ ] Decidir se haverá push externo, além das notificações internas.
- [ ] Avaliar automatização do reprocessamento de embeddings e índices vetoriais quando o volume e a operação justificarem.
