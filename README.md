# eco-knock-be-embedded

라즈베리파이에서 `BME680` 센서를 읽고, 읽은 값을 중앙 백엔드로 `gRPC` 전송하는 임베디드 서버입니다.

이 프로젝트는 아직 개발 중입니다.  
현재 README는 **지금 구현된 범위만** 설명하며, 이후 기능이 추가될 때마다 함께 갱신하는 것을 전제로 합니다.

## 현재 구현 범위

현재 구현된 기능은 아래와 같습니다.

- `BME680` raw 값 읽기
  - `temperature`
  - `humidity`
  - `gas_resistance`
- 3초 주기 센서 polling
- 중앙 백엔드로 `gRPC` 전송
- Raspberry Pi용 Docker 배포

아직 문서화되지 않은 기능이나 운영 흐름은 이후 구현에 맞춰 정리합니다.

## 디렉토리 구조

주요 디렉토리는 아래 정도만 보면 됩니다.

- `cmd/server`
  - 애플리케이션 부팅
  - 센서 리포터 wiring
- `internal/sensor/bme680`
  - BME680 센서 접근
- `internal/sensor/streaming`
  - 주기 polling 후 stream 생성
- `internal/sensor/report`
  - stream을 받아 중앙 백엔드로 전송
- `internal/grpc`
  - gRPC client
  - generated protobuf code
- `proto`
  - `.proto` 원본
- `scripts`
  - proto 생성
  - Raspberry Pi 배포 스크립트

## 설정 방식

설정은 `application.yaml`이 읽고, 실제 값은 `.env`에서 주입합니다.

`application.yaml`은 `${ENV_NAME}` 형식으로 환경변수를 참조합니다.

현재 필요한 값은 아래와 같습니다.

```env
SERVER_HTTP_PORT=19090

CENTRAL_BACKEND_HOST=192.168.0.11
CENTRAL_BACKEND_HTTP_PORT=18080
CENTRAL_BACKEND_GRPC_PORT=6565
ALLOW_CENTRAL_BACKEND_FAILURE=true

SENSOR_I2C_DEVICE=/dev/i2c-1
SENSOR_I2C_ADDRESS=0x76
```

설정 의미:

- `SERVER_HTTP_PORT`
  - 이 서버가 listen 하는 HTTP 포트
- `CENTRAL_BACKEND_HOST`
  - 센서 데이터를 받을 중앙 백엔드 주소
- `CENTRAL_BACKEND_GRPC_PORT`
  - 중앙 백엔드 gRPC 포트
- `ALLOW_CENTRAL_BACKEND_FAILURE`
  - 중앙 백엔드 연결 실패 시 서버를 계속 올릴지 여부
- `SENSOR_I2C_DEVICE`
  - Raspberry Pi의 I2C 디바이스 경로
- `SENSOR_I2C_ADDRESS`
  - BME680 주소 (`0x76` 또는 `0x77`)

## 로컬 실행

로컬 개발 환경에서는 아래 순서로 실행합니다.

1. `.env` 준비
2. 서버 실행

```powershell
go run ./cmd/server
```

테스트:

```powershell
go test ./...
```

주의:

- Windows에서는 실제 BME680 대신 stub reader가 동작합니다.
- Linux, 특히 Raspberry Pi에서는 `/dev/i2c-1`와 실제 센서가 필요합니다.

## Proto

proto 원본은 아래에 있습니다.

- [`proto/sensor/v1/sensor.proto`](./proto/sensor/v1/sensor.proto)

generated code를 다시 만들 때는 아래 명령을 사용합니다.

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\gen-proto.ps1
```

## Raspberry Pi Docker 배포

이 프로젝트는 Raspberry Pi에서 직접 build 하지 않고, 개발 머신에서 이미지를 빌드해서 Pi에 올리는 흐름을 기준으로 합니다.

필요 파일:

- `.env`
- `.env.deploy`
- `docker-compose.yml`

예시 `.env.deploy`:

```env
PI_HOST=192.168.0.28
PI_USER=pi
PI_SSH_PORT=22
PI_APP_DIR=~/eco-knock-be-embedded
IMAGE_NAME=eco-knock-be-embedded:arm64
DOCKER_PLATFORM=linux/arm64
```

Windows에서 배포:

```powershell
.\scripts\deploy-pi.ps1
```

Shell 환경에서 배포:

```bash
./scripts/deploy-pi.sh
```

현재 Docker 배포는 `linux/arm64` 기준입니다.  
Raspberry Pi에서 `uname -m` 결과가 `aarch64`인지 먼저 확인해야 합니다.

## Raspberry Pi 점검

현재 기준으로 BME680은 Raspberry Pi GPIO I2C 버스 1 기준으로 사용합니다.

- `SENSOR_I2C_DEVICE=/dev/i2c-1`
- `SENSOR_I2C_ADDRESS=0x76`

배포 후 확인 명령:

```bash
ls -l /dev/i2c*
sudo i2cdetect -y 1
docker ps
cd ~/eco-knock-be-embedded
docker compose logs --tail=100
```

정상이라면 `i2cdetect -y 1`에서 `76`이 보여야 합니다.

## 로컬 스킬

이 저장소에는 로컬 스킬이 있습니다.

- `git-commit-korean`
  - 한국어 커밋 메시지 규칙과 작은 단위 커밋 분할
- `raspberry-pi-docker-deploy`
  - Raspberry Pi Docker 배포와 I2C/BME680 점검

스킬 레지스트리는 [`AGENTS.md`](./AGENTS.md)에서 관리합니다.

## 현재 제약

현재는 아래를 전제로 합니다.

- 센서 1개
- `DeviceId = 1` 고정
- 중앙 백엔드 gRPC 계약은 `sensor.v1.SensorService` 기준
- 운영 구성이 계속 바뀔 수 있으므로 README도 기능 추가 때마다 갱신 예정
