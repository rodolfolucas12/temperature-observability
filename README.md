# SISTEMA DE CONSULTA DE TEMPERATURA POR CEP COM OBSERVABILIDADE

 - service_input: Recebe um CEP, valida e encaminha ao service_orchestrator.


 - service_orchestrator: Consulta a localização e o clima do CEP e retorna as temperaturas em Celsius, Fahrenheit e Kelvin.


# Tecnologias Utilizadas
- Go
- Docker/Docker Compose
- OpenTelemetry (OTEL)
- Zipkin

# Executando o Projeto

### Para executar o projeto, siga os passos abaixo:
- Clone o repositório: 
    git clone https://github.com/rodolfolucas12/temperature-observability

- execute o docker-compose:
    docker compose up --build

- Acesse o service_input em http://localhost:8080/cep

- Exemplo de Requisição para testar a resposta

curl -X POST http://localhost:8080/cep -d '{"cep": "02228010"}' -H "Content-Type: application/json"

- Postman
    POST http://localhost:8080/cep

    Body:
    ```json
    {
        "cep": "02228010"
    }
    ```
- Observabilidade com OTEL e Zipkin
    Acesse a url http://localhost:9411
