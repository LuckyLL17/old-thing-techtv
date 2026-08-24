# Bug 复现说明

## Bug 是什么

审计统计接口选择最近 N 天时，较早日期的旧日志仍被计算进去，时间范围越短，结果与明细的差异越明显；请定位审计统计的时间窗口和分组统计之间的链路问题，说明原因即可，不要修改代码。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/tool.go、internal/repository/extras_repo.go、internal/service/extras_service.go。

## 根因

1、涉及生产文件：internal/domain/tool.go、internal/repository/extras_repo.go、internal/service/extras_service.go；关键生产符号：ExtrasService.StatsByAction、ExtrasRepo.StatsByAction、domain.AuditLog.CreatedAt。
2、统计窗口计算与分组查询使用了不一致的时间边界/日期方向，统计链路没有复用明细接口的有效 from 条件，短窗口下旧日志仍被纳入或统计结果为空。
3、失效机制属于状态一致性与跨层数据流：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug025VerificationAuditStatsWindow$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug025VerificationAuditStatsRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug025VerificationAuditStatsWindow
    bug025_verification_test.go:40: StatsByAction days window mismatch: created_at filter returned map[]
--- FAIL: TestBug025VerificationAuditStatsWindow (0.00s)
FAIL
FAIL	upcycle-hub/internal/verification	0.492s
FAIL
```
