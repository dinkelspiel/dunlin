backend:
    docker compose up -d mariadb
    sh -c 'until docker compose exec -T mariadb mariadb-admin ping -ucdn -pcdn --silent >/dev/null 2>&1; do echo "Waiting for MariaDB..."; sleep 1; done'
    MARIADB_DATABASE_URL='cdn:cdn@tcp(127.0.0.1:3306)/cdn' STORAGE_URL='./.data/dunlin' HOST_ROOT='.' APP_URL='localhost' watchexec --restart -- 'go build -o main . && ./main'

frontend:
    cd frontend && VITE_API_URL='http://localhost:8080' pnpm dev