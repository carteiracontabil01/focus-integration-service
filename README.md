# Focus Integration Service

Microservice em Go para integrar a **Carteira Contábil** com a **Focus NFe**, focado em **cadastro/edição/consulta/exclusão de empresas** (`/v2/empresas`).

Referência: [Documentação oficial - Empresas](https://focusnfe.com.br/doc/#empresas)

---

## Arquitetura

```
focus-integration-service/
├── cmd/api/           # entrypoint
├── internal/
│   ├── config/        # env
│   ├── focus/         # http client Focus
│   ├── handler/       # http handlers (REST)
│   └── server/        # router + middlewares
├── docs/              # swagger (swag)
├── Makefile
├── go.mod / go.sum
└── env.example
```

---

## Variáveis de ambiente

Veja `env.example`.

---

## Rodando

```bash
make run
```

Health:

`GET /health`

Swagger UI:

`GET /swagger/index.html`

---

## Endpoints (proxy Focus)

- `POST   /v2/empresas`
- `GET    /v2/empresas`
- `GET    /v2/empresas/{id}`
- `PUT    /v2/empresas/{id}`
- `DELETE /v2/empresas/{id}`

## Endpoints (consulta de CNPJ)

- `GET    /v2/cnpjs/{cnpj}` (14 dígitos, somente números)

## Endpoints (municípios - beta)

- `GET    /v2/municipios`
- `GET    /v2/municipios/{codigo_municipio}`
- `GET    /v2/municipios/{codigo_municipio}/itens_lista_servico`
- `GET    /v2/municipios/{codigo_municipio}/itens_lista_servico/{codigo}`
- `GET    /v2/municipios/{codigo_municipio}/codigos_tributarios_municipio`
- `GET    /v2/municipios/{codigo_municipio}/codigos_tributarios_municipio/{codigo}`


