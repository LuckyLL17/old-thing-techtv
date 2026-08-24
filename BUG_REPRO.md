# Bug 复现说明

## Bug 是什么

教程后台筛选 published 时，列表里仍会出现 draft，统计数字也跟着把两种状态混在一起；单独查看详情又能看到正确状态，说明问题发生在教程后台的列表查询链路中，请定位状态条件失效的位置，不要修改代码。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/tutorial.go、internal/repository/tutorial_repo.go、internal/service/tutorial_service.go。

## 根因

1、涉及生产文件：internal/domain/tutorial.go、internal/repository/tutorial_repo.go、internal/service/tutorial_service.go；关键生产符号：TutorialService.List、TutorialRepo.List、domain.Tutorial.Status。
2、后台列表链路没有把 published 状态条件稳定传到 repository 查询，列表与统计走了混合状态的数据集，而详情读取使用了另一条正确路径。
3、失效机制属于跨层数据流：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug016VerificationTutorialStatusFilter$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug016VerificationTutorialStatusFilterRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug016VerificationTutorialStatusFilter
    bug016_verification_test.go:40: status filter total=2 list=[0xd5ac4459a20 0xd5ac4459b80]
--- FAIL: TestBug016VerificationTutorialStatusFilter (0.01s)
FAIL
FAIL	upcycle-hub/internal/verification	0.458s
FAIL
```
