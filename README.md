## ✏️ Introdução
Minha implementação do projeto de Go da NLW disponibilizado pela Rocketseat. Uma api que permita cadastro e edição de viagens, convidar pessoas, adicionar atividades e links para a sua viagem. O projeto está dividido de acordo com as aulas, cada branch representa uma aula, além de uma branch de conteúdos que não foran mostrados na aula, eu mesmo desenvolvi com os conhecimentos que aprendi no curso.

## 📋 Features adicionais
- Na NLW, somente 2 endpoints foram feito, já na minha implementação todos os endpoints estão feitos e funcionando.

## 🚧 Rotas
  - Trips
    - POST /trips - Cria uma nova viagem e manda um e-mail placeholder de confirmação
    - GET /trips/{tripId} - Retorna os detalhes de uma viagem
    - PUT /trips/{tripId} - Atualiza uma viagem
    - GET /trips/{tripId}/confirm - Confirma uma viagem
  - Participants
    - POST /trips/{tripId}/invites - Cria um novo participante para uma viagem
    - GET /trips/{tripId}/participants - Retorna todos os participantes de uma viagem
    - PATCH /participants/{participantId}/confirm - Confirma uma participante em uma viagem
  - Activities
    - POST /trips/{tripId}/activities - Cria uma nova atividade para uma viagem
    - GET /trips/{tripId}/activities - Retorna todas as atividades de uma viagem
  - Links
    - POST /trips/{tripId}/links - Cria um novo link para uma viagem
    - GET /trips/{tripId}/links - Retorna todos os links de uma viagem

## 💻 Tecnologias

* [Go](https://go.dev/)
* [Docker](https://www.docker.com/)
* [Postgres](https://www.postgresql.org/)
* [Mailpit](https://mailpit.axllent.org/)
