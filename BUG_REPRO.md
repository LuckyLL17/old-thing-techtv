# Bug 复现说明

## Bug 是什么

学习记录超过一页后，用户翻到第二页会再次看到第一页末尾那条记录，返回总数却与实际去重后的结果不相符；这会让用户误以为有重复学习记录，请调整分页处理，使页间边界和总数保持一致。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/attempt.go、internal/repository/attempt_repo.go、internal/service/interaction_service.go。

## 根因

1、涉及生产文件：internal/domain/attempt.go、internal/repository/attempt_repo.go、internal/service/interaction_service.go；关键生产符号：AttemptRepo.ListByUser、InteractionService.ListAttempts、domain.Attempt。
2、学习记录分页使用了错误的 offset 基准，第二页的起点没有按同一页码语义推进，结果与 total 的统计口径不一致并重复上一页末尾记录。
3、失效机制属于状态一致性与跨层数据流：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug015VerificationAttemptPagination$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug015VerificationAttemptPaginationRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug015VerificationAttemptPagination
    bug015_verification_test.go:44: pages overlap total=4 a=[0x5c52fadd9260 0x5c52fadd92d0] b=[]
--- FAIL: TestBug015VerificationAttemptPagination (0.00s)
FAIL
FAIL	upcycle-hub/internal/verification	0.445s
FAIL
```
