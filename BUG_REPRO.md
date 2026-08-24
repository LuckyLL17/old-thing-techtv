# Bug 复现说明

## Bug 是什么

有用户因为网络抖动连续点击注册，或者客户端把同一个请求重放了几次，结果系统里出现多个相同邮箱的账号；接口日志出现“duplicate email accepted --- FAIL”，请调整注册链路，让邮箱唯一性在重试场景下也能保持。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/user.go、internal/repository/user_repo.go、internal/service/auth_service.go。

## 根因

1、涉及生产文件：internal/domain/user.go、internal/repository/user_repo.go、internal/service/auth_service.go；关键生产符号：AuthService.Register、UserRepo.GetByEmail、UserRepo.Create、domain.User.Email。
2、注册预检查使用了错误的查询列且重复命中后未阻断，同时模型和实际迁移缺少可靠的邮箱唯一约束；网络重放或并发请求因此能穿过预检查并写入重复账号。
3、失效机制属于并发与状态一致性：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug003VerificationDuplicateRegistration$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug003VerificationDuplicateRegistrationRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug003VerificationDuplicateRegistration
.../env/internal/repository/user_repo.go:44 record not found
[0.021ms] [rows:0] SELECT * FROM `users` WHERE username = "same@example.com" ORDER BY `users`.`id` LIMIT 1
.../env/internal/repository/user_repo.go:56 record not found
[0.013ms] [rows:0] SELECT * FROM `users` WHERE username = "alice" ORDER BY `users`.`id` LIMIT 1
.../env/internal/repository/user_repo.go:44 record not found
[0.362ms] [rows:0] SELECT * FROM `users` WHERE username = "same@example.com" ORDER BY `users`.`id` LIMIT 1
.../env/internal/repository/user_repo.go:56 record not found
[0.054ms] [rows:0] SELECT * FROM `users` WHERE username = "alice2" ORDER BY `users`.`id` LIMIT 1
    bug003_verification_test.go:38: duplicate email accepted
--- FAIL: TestBug003VerificationDuplicateRegistration (0.11s)
FAIL
FAIL	upcycle-hub/internal/verification	1.028s
FAIL
```
