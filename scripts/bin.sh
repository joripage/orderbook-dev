# script

SCRIPTPATH="$(
    cd "$(dirname "$0")"
    pwd -P
)"

CURRENT_DIR=$SCRIPTPATH
ROOT_DIR="$(dirname $CURRENT_DIR)"

INFRA_LOCAL_COMPOSE_FILE=$ROOT_DIR/build/docker-compose.dev.yaml

function local_infra() {
    docker-compose -f $INFRA_LOCAL_COMPOSE_FILE $@
}

function infra() {
    case $1 in
        up)
            local_infra up ${@:2}
        ;;
        down)
            local_infra down ${@:2}
        ;;
        build)
            local_infra build ${@:2}
        ;;
        *)
            echo "up|down|build [docker-compose command arguments]"
        ;;
    esac
}

# Setup variables environment for app
function setup_env_variables() {
    set -a
    export $(grep -v '^#' "$ROOT_DIR/build/.base.env" | xargs -0) >/dev/null 2>&1
    . $ROOT_DIR/build/.base.env
    set +a
    export CONFIG_FILE=$ROOT_DIR/build/config.yaml
}

# Run test
function run_test() {
    # run all unit tests
    echo 'run unit testing'
    go test ./... -short -cover -coverprofile=coverage.out || {
        echo 'unit testing failed'
        exit 1
    }
}

# Open test coverage html
function open_test_coverage() {
    go tool cover -html coverage.out
}

# Lint
function lint() {
    LINT_PATH="$ROOT_DIR/bin/golangci-lint"
    $LINT_PATH run --timeout 5m
}

# Start app
function start_service() {
    infra up -d
    setup_env_variables
    ENTRY_FILE="$ROOT_DIR/cmd/oms/main.go"
    
    echo "Starting app with config file: $CONFIG_FILE"
    go run $ENTRY_FILE --config-file=$CONFIG_FILE
}

# Start infra
function start_infra() {
    infra up -d
}

function migrate() {
    echo "Starting migration..."
    infra up -d
    setup_env_variables
    ENTRY_FILE="$ROOT_DIR/cmd/migrate/main.go"
    go run $ENTRY_FILE
}

# Main
case $1 in
    start_service)
        start_service
    ;;
    lint)
        lint
    ;;
    test)
        run_test
    ;;
    test_cover)
        open_test_coverage
    ;;
    start_infra)
        start_infra
    ;;
    migrate)
        migrate 
    ;;
    *)
        echo $0 "[start_service|lint|test|test_cover|start_infra|migrate]"
    ;;
esac