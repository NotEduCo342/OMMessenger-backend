# Docker Swarm vs Docker Compose: Cheat Sheet

Docker Swarm uses "Stacks" and "Services" instead of raw containers. A **Stack** is a collection of services (defined by your `docker-compose.yml`), and a **Service** is the blueprint for running one or more identical container **Tasks** (replicas).

Here is a quick reference guide comparing commands you are used to in `docker-compose` with their `docker swarm` / `docker stack` equivalents.

## 🚀 1. Starting / Deploying

**Docker Compose:**
```bash
docker-compose up -d
```

**Docker Swarm:**
```bash
docker stack deploy -c docker-compose.yml <stack_name>
# Example: docker stack deploy -c deploy/dev/docker-compose.yml om-dev
```
> [!TIP]
> `docker stack deploy` automatically creates the services or updates them if they already exist. There is no separate "update" command.

---

## 🛑 2. Stopping / Tearing Down

**Docker Compose:**
```bash
docker-compose down
```

**Docker Swarm:**
```bash
docker stack rm <stack_name>
# Example: docker stack rm om-dev
```
> [!WARNING]
> This removes the stack and all its services, networks, and containers immediately. Volumes are preserved.

---

## 📋 3. Viewing Running Services (PS)

**Docker Compose:**
```bash
docker-compose ps
```

**Docker Swarm (Stack Level):**
```bash
docker stack services <stack_name>
# Example: docker stack services om-dev
```
*Lists all services inside a specific stack and shows how many replicas are running.*

**Docker Swarm (Global Level):**
```bash
docker service ls
```
*Lists absolutely every service running on your entire Swarm cluster.*

---

## 🪵 4. Viewing Logs

**Docker Compose:**
```bash
docker-compose logs -f backend
```

**Docker Swarm:**
```bash
docker service logs -f <stack_name>_<service_name>
# Example: docker service logs -f om-dev_backend
```
> [!NOTE]
> In Swarm, the service name is prefixed by the stack name. It will aggregate logs from *all* replicas of that service across all servers.

---

## 🔍 5. Inspecting Specific Containers (Tasks)

If a service is failing or restarting, you'll want to see its specific history. A "Task" in Swarm is an individual container.

**Docker Swarm:**
```bash
docker service ps <stack_name>_<service_name>
# Example: docker service ps om-dev_postgres
```
*This shows the history of the container, what node it is running on, and exactly why it might have failed or been rejected.*

---

## ⚖️ 6. Scaling Services

**Docker Compose:**
```bash
docker-compose up -d --scale backend=3
```

**Docker Swarm:**
```bash
docker service scale <stack_name>_<service_name>=<number>
# Example: docker service scale om-dev_backend=3
```
*Swarm will instantly spin up new replicas and add them to the internal load balancer.*

---

## 🔄 7. Force Updating / Restarting a Service

**Docker Compose:**
```bash
docker-compose restart backend
```

**Docker Swarm:**
```bash
docker service update --force <stack_name>_<service_name>
# Example: docker service update --force om-dev_backend
```
*This forces Swarm to gracefully kill the existing tasks and start fresh ones, pulling the latest image if necessary.*

---

## 🐚 8. Executing Commands Inside a Container

Because Swarm can run containers on multiple servers, there is no direct `docker service exec`. You must first find the specific container ID running on your local node.

**Docker Compose:**
```bash
docker-compose exec backend sh
```

**Docker Swarm (2 Steps):**
```bash
# 1. Find the container ID on your machine
docker ps | grep om-dev_backend

# 2. Exec into it exactly like a normal container
docker exec -it <container_id> sh
```
