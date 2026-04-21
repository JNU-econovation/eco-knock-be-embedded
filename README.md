# eco-knock-be-embedded

라즈베리파이에서 센서와 공기청정기 상태를 읽고, 외부 서버가 `gRPC`로 현재 상태를 조회할 수 있게 제공하는 임베디드 서버입니다.

이 프로젝트는 아직 개발 중입니다. README는 현재 구현된 범위만 설명하고, 기능이 추가될 때마다 갱신하는 것을 전제로 합니다.

## 현재 구현 범위

- `BME680` 현재값 1회 조회
  - `sensor.v1`
    - 온도
    - 습도
    - 가스 저항
    - 센서 상태 비트
  - `sensor.v2`
    - `sensor.v1`의 raw 값
    - `static_iaq`
    - `estimated_eco2_ppm`
    - `estimated_bvoc_ppm`
    - `accuracy`
    - `stabilization_progress_pct`
    - `gas_percentage`
    - `learning_complete_at_unix_ms`
- 샤오미 공기청정기 2 계열 현재 상태 조회
  - 전원
  - AQI
  - 평균 AQI
  - 습도
  - 온도
  - 모드
  - favorite level
  - 필터 수명
  - 팬 속도
  - LED, 부저, 차일드락
- 라즈베리파이에서 `gRPC` 서버 실행
- Prometheus scrape용 `GET /metrics` 제공
- 개발 PC에서 Docker 이미지를 빌드하고 Raspberry Pi에서 실행하는 배포 흐름

현재 기준으로 중앙 서버가 이 서버를 조회하는 주 인터페이스는 `gRPC`입니다.

## 주요 디렉토리

- `cmd/server`
  - 서버 부팅과 REST, gRPC 서버 wiring
- `internal/sensor`
  - BME680 읽기와 센서 조회 서비스
- `internal/airpurifier/xiaomi`
  - 샤오미 공기청정기 miIO 통신
- `internal/grpc/server`
  - 센서, 공기청정기 gRPC 서버 어댑터
- `internal/common`
  - 공통 설정, 에러, 미들웨어
- `proto`
  - gRPC 계약
- `scripts`
  - proto 생성, Raspberry Pi 배포 스크립트

## 설정 방식

설정은 `application.yaml`을 읽고, 실제 값은 `.env`에서 주입합니다.

`application.yaml`:

```yaml
server:
  http_port: ${SERVER_HTTP_PORT}
  grpc_port: ${SERVER_GRPC_PORT}

central_backend:
  host: ${CENTRAL_BACKEND_HOST}
  http_port: ${CENTRAL_BACKEND_HTTP_PORT}
  grpc_port: ${CENTRAL_BACKEND_GRPC_PORT}
  allow_failure: ${ALLOW_CENTRAL_BACKEND_FAILURE}

sensor:
  i2c_device: ${SENSOR_I2C_DEVICE}
  i2c_address: ${SENSOR_I2C_ADDRESS}
  state_db_path: data/sensor_state.db
  heater_temp_c: 300
  heater_duration: 100ms
  ambient_temp_c: 25
  poll_interval: 1s
  state_checkpoint_valid_samples: 60
  air_quality:
    history_limit: 3600
    stabilization_duration: 5m
    learning_duration: 1h
    stabilization_valid_sample_goal: 300
    learning_valid_sample_goal: 3600

air_purifier:
  address: ${AIR_PURIFIER_ADDRESS}
  token: ${AIR_PURIFIER_TOKEN}
  timeout: ${AIR_PURIFIER_TIMEOUT}
```

`.env` 예시:

```env
SERVER_HTTP_PORT=19090
SERVER_GRPC_PORT=6565
NODE_EXPORTER_HTTP_PORT=9100

CENTRAL_BACKEND_HOST=192.168.0.11
CENTRAL_BACKEND_HTTP_PORT=18080
CENTRAL_BACKEND_GRPC_PORT=6565
ALLOW_CENTRAL_BACKEND_FAILURE=true

SENSOR_I2C_DEVICE=/dev/i2c-1
SENSOR_I2C_ADDRESS=0x76

AIR_PURIFIER_ADDRESS=192.168.0.50:54321
AIR_PURIFIER_TOKEN=0123456789abcdef0123456789abcdef
AIR_PURIFIER_TIMEOUT=3s
```

설명:

- `SERVER_HTTP_PORT`
  - Gin HTTP 서버 포트
- `SERVER_GRPC_PORT`
  - 외부 서버가 상태 조회에 사용하는 gRPC 포트
- `NODE_EXPORTER_HTTP_PORT`
  - node exporter가 Raspberry Pi 호스트 CPU, 메모리, 네트워크, 디스크 I/O metrics를 노출하는 포트
- `SENSOR_I2C_DEVICE`
  - Raspberry Pi의 I2C 디바이스 경로
- `SENSOR_I2C_ADDRESS`
  - BME680 주소
- `sensor.state_db_path`
  - air quality 학습 상태를 저장하는 SQLite 경로
- `sensor.heater_temp_c`, `sensor.heater_duration`, `sensor.ambient_temp_c`
  - BME680 강제 측정 히터 설정
- `sensor.poll_interval`
  - 센서 백그라운드 폴링 주기
- `sensor.state_checkpoint_valid_samples`
  - 유효 샘플 몇 개마다 SQLite 상태를 저장할지 결정
- `sensor.air_quality.*`
  - `sensor.v2` 파생값 계산에 사용하는 history, stabilization, learning 설정
- `AIR_PURIFIER_ADDRESS`
  - 샤오미 공기청정기 로컬 miIO 주소
- `AIR_PURIFIER_TOKEN`
  - 32자리 miIO 토큰
- `AIR_PURIFIER_TIMEOUT`
  - 공기청정기 요청 타임아웃

`central_backend` 관련 값은 현재 outbound 연동이 아니라 설정 호환성을 위해 남아 있는 예약 필드입니다.

## 로컬 실행

```powershell
go run ./cmd/server
```

테스트:

```powershell
go test ./...
```

현재 Gin HTTP 서버는 실행되지만, 비즈니스 HTTP 엔드포인트는 아직 없습니다. 실제 상태 조회 인터페이스는 gRPC입니다.

Prometheus metrics:

```powershell
curl http://localhost:19090/metrics
```

애플리케이션 metrics는 `GET /metrics`에서 노출합니다. Raspberry Pi 호스트 리소스 metrics는 compose의 `node-exporter` 서비스가 `NODE_EXPORTER_HTTP_PORT`로 노출합니다.

## Docker Compose

개발용 compose:

```powershell
docker compose up --build
```

개발용 `docker-compose.yml`은 다음 서비스를 실행합니다.

- `node-exporter`
  - Windows 개발 환경에서 `NODE_EXPORTER_HTTP_PORT`를 publish해서 `/metrics` 확인

개발 환경에서 앱 서버까지 실행할 때는 별도로 `go run ./cmd/server`를 사용합니다. Raspberry Pi용 compose 파일은 `docker-compose.pi.yml`이며, Pi에서는 node exporter를 host network로 실행합니다.

## gRPC 계약

현재 구현된 RPC:

- `sensor.v1.SensorService/GetCurrentSensor`
- `sensor.v2.SensorService/GetCurrentSensor`
- `airpurifier.v1.AirPurifierService/GetCurrentAirPurifier`

proto 파일:

- [sensor.v1](./proto/sensor/v1/sensor.proto)
- [sensor.v2](./proto/sensor/v2/sensor.proto)
- [airpurifier.proto](./proto/airpurifier/v1/airpurifier.proto)

센서 버전 구분:

- `sensor.v1`
  - raw 센서값만 제공합니다.
- `sensor.v2`
  - raw 센서값에 air quality 파생값을 함께 제공합니다.

현재 공기청정기는 조회만 구현돼 있고, 전원 제어나 모드 변경 RPC는 아직 없습니다.

## Proto 생성

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\gen-proto.ps1
```

이 스크립트는 `proto/` 아래의 `.proto` 파일을 순회하면서 Go protobuf, gRPC 코드를 다시 생성합니다.

## Raspberry Pi 배포

현재 기본 배포 흐름은 Raspberry Pi에서 직접 빌드하지 않고, 개발 PC에서 이미지를 빌드해서 Pi에 올리는 방식입니다.

PowerShell:

```powershell
.\scripts\deploy-pi.ps1
```

Shell:

```bash
./scripts/deploy-pi.sh
```

배포 스크립트가 하는 일:

1. 개발 PC에서 `linux/arm64` 이미지 빌드
2. Pi에 `docker-compose.pi.yml`, `.env` 전송
3. Docker 이미지를 `docker load`로 Pi에 적재
4. Pi에서 `docker compose up -d`

배포 스크립트는 `docker-compose.pi.yml`을 Pi의 배포 디렉토리에 `docker-compose.yml` 이름으로 전송합니다.

`docker-compose.pi.yml`은 현재 다음 특징을 가집니다.

- `linux/arm64` 기준
- `/dev/i2c-1` 디바이스 마운트
- 컨테이너를 root로 실행
- 앱 HTTP 포트 publish
- node exporter를 host network로 실행해서 Pi 호스트 metrics 노출

즉 Pi용 compose는 `SERVER_HTTP_PORT`를 publish하고, node exporter는 host network에서 `NODE_EXPORTER_HTTP_PORT`로 노출합니다. `SERVER_GRPC_PORT`는 publish하지 않습니다.

## Raspberry Pi 센서 연결

BME680은 현재 다음 기준으로 동작합니다.

- I2C 디바이스: `/dev/i2c-1`
- 주소: `0x76`
- 백그라운드 폴링: 기본 `1s`
- 상태 저장: SQLite checkpoint, 기본 `60`개 유효 샘플마다 저장

Pi에서 점검할 때 자주 쓰는 명령:

```bash
ls -l /dev/i2c*
sudo i2cdetect -y 1
docker ps
docker compose logs --tail=100
```

정상 인식이면 `i2cdetect -y 1` 결과에 `76`이 보여야 합니다.

## 현재 제약 사항

- HTTP 비즈니스 API는 아직 없습니다.
- Prometheus용 HTTP 엔드포인트는 `GET /metrics`만 제공합니다.
- 공기청정기 gRPC는 현재 조회만 있고 제어 RPC는 없습니다.
- `central_backend` 설정은 남아 있지만 현재 주 경로는 `Spring -> Go gRPC 조회`입니다.
- Pi용 Docker compose는 현재 앱 HTTP 포트만 publish하고, node exporter는 host network로 노출합니다.
- 샤오미 공기청정기 miIO 토큰은 자동 추출하지 않습니다.
- `sensor.v2`의 `static_iaq`, `estimated_eco2_ppm`, `estimated_bvoc_ppm`는 BSEC 출력이 아니라 서버 내부 추정값입니다.
- non-linux 환경에서는 공기청정기 gRPC가 stub 클라이언트로 동작하고, Linux에서만 실제 miIO 클라이언트를 사용합니다.

## 관련 스킬

이 저장소에는 로컬 스킬이 있습니다.

- `git-commit-korean`
- `raspberry-pi-docker-deploy`
- `readme-maintainer`
- `eco-knock-maintainer`

스킬 목록과 등록 상태는 [AGENTS.md](./AGENTS.md)에서 관리합니다.
