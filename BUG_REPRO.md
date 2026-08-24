# Bug 复现说明

## Bug 是什么

运营人员将被禁用的账号标记为停用后，该用户仍然可以用原来的凭据登录并拿到令牌，提示报错“disabled account received a token --- FAIL”；请定位登录状态校验为何没有阻断这条认证链路而导致报错，不要修改代码。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/user.go、internal/repository/user_repo.go、internal/service/auth_service.go。

## 根因

1、涉及生产文件：internal/domain/user.go、internal/repository/user_repo.go、internal/service/auth_service.go；关键生产符号：AuthService.Login、AuthService.GenerateToken、domain.User.Status。
2、Login 命中非启用状态后只执行空操作，没有在 GenerateToken 前返回禁用错误；后续密码校验和令牌签发继续执行，账号状态没有真正阻断认证链路。
3、失效机制属于跨层数据流与状态一致性：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug004VerificationDisabledAccountCannotLogin$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug004VerificationDisabledAccountRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug004VerificationDisabledAccountCannotLogin
    bug004_verification_test.go:50: disabled account received a token
--- FAIL: TestBug004VerificationDisabledAccountCannotLogin (0.01s)
FAIL
FAIL	upcycle-hub/internal/verification	0.465s
FAIL
```
