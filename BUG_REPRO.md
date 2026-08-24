# Bug 复现说明

## Bug 是什么

我的收藏页请求了 tutorial 类型，想只查看教程收藏，结果列表中却混进了 project；复现时提示“target_type filter mismatch: expected tutorial favorites, total=2 --- FAIL”，请查清类型条件是在请求传递、查询还是组装结果时丢掉的，不要改代码。

## 如何触发

在与题面相同的业务场景中准备最小数据，执行下面的目标验证命令，观察认证、列表、统计或事务结果。该现象对应的生产调用链涉及：internal/domain/favorite.go、internal/repository/favorite_repo.go、internal/service/interaction_service.go。

## 根因

1、涉及生产文件：internal/domain/favorite.go、internal/repository/favorite_repo.go、internal/service/interaction_service.go；关键生产符号：InteractionService.ListFavorites、FavoriteRepo.ListByUser、domain.Favorite.TargetType。
2、handler 传入的 targetType 在 service 被替换为空串，repository 即便进入过滤分支也没有按 target_type 查询，导致 tutorial 与 project 收藏一起返回。
3、失效机制属于跨层数据流：跨层调用中的参数、状态或资源边界没有保持一致，导致接口返回与持久化结果出现偏差。

## 运行指令

```bash
go test command 1: go test ./internal/verification -v -run '^TestBug006VerificationFavoriteTypeFilter$' -count=1
go test command 2: go test ./internal/verification -v -run '^TestBug006VerificationFavoriteTypeFilterRegression$' -count=1
```

## 错误信息

目标场景在修复前会出现题面描述的失败现象；回归场景用于确认基础调用链仍可运行。

## 错误堆栈

```text
=== RUN   TestBug006VerificationFavoriteTypeFilter
    bug006_verification_test.go:42: target_type filter mismatch: expected tutorial favorites, total=2 list=[0x4e7635dd2a40 0x4e7635dd2ac0]
--- FAIL: TestBug006VerificationFavoriteTypeFilter (0.00s)
FAIL
FAIL	upcycle-hub/internal/verification	0.443s
FAIL
```
