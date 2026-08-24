# old-thing-techtv

项目用途：面向手工爱好者与环保生活践行者的创意分享平台。用户可发布将废旧物品改造成新用品的教程（如旧牛仔裤改背包、玻璃瓶改台灯、木托盘改花架），其他人可浏览、收藏、评论、评分，并按教程复刻改造。理念：**让旧物重新发光**。项目源代码、依赖描述和评测专用 Docker 文件共同构成自包含任务；不依赖本机预编译二进制。

## 标准构建、运行和测试命令

```bash
go build ./...
go run ./cmd/server
go test ./...
```
## 评测容器

评测专用 Dockerfile 为 `benzhi.Dockerfile`，构建脚本为 `build_benzhi_docker.sh`。

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```
