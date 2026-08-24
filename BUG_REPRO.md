# Bug 复现说明

## Bug 是什么

用户给同一篇教程提交重复标签时，保存过程会在标签关联表处撞上重复主键，结果不仅标签没有正常保存，整次教程保存事务也被回滚；请定位标签 upsert 与关联写入之间如何造成状态污染，代码不要修改。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/tag.go、internal/repository/tag_repo.go、internal/service/category_service.go。

## 根因

1、涉及生产文件：internal/domain/tag.go、internal/repository/tag_repo.go、internal/service/category_service.go；关键生产符号：CategoryService.Upsert、TagRepo.UpsertByName、TagRepo.LinkTutorial、TutorialService.Create。
2、重复标签名在 upsert 返回切片中没有去重，关联表第二次插入同一联合主键失败；关联子事务回滚却没有覆盖前面已提交的教程及其他实体写入，形成半成功状态。
3、失效机制属于状态污染与跨层数据流：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug030VerificationDuplicateTags$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug030VerificationDuplicateTagsRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug030VerificationDuplicateTags
.../env/internal/repository/tag_repo.go:44 record not found
[0.015ms] [rows:0] SELECT * FROM `tags` WHERE name = "wood" ORDER BY `tags`.`id` LIMIT 1
    bug030_verification_test.go:37: upsert returned 2 tags
--- FAIL: TestBug030VerificationDuplicateTags (0.00s)
FAIL
FAIL	upcycle-hub/internal/verification	0.433s
FAIL
```
