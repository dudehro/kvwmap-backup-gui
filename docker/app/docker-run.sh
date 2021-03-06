#docker rm -f kvwmap-backup
echo "Updating images..."
docker pull golang:alpine
docker pull alpine:latest
echo "Starting image build..."
docker build --no-cache -t gkaemmert/kvwmap-backup:1.0 .
echo "Starting container in foreground..."
docker run --rm  -p 8500:8500 \
-v /home/work/github.com/kvwmap-backup/backup-config/:/app/backup-config \
--name kvwmap-backup gkaemmert/kvwmap-backup:1.0
