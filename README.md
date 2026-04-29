# agent-starter

`agent-starter` 提供基于 `gocfg-load-module` 的模块化启动样板，强调显式依赖、稳定加载顺序和完整生命周期。

## 目录

```text
agent-starter/
  modules/
    db_module.go
    db_module_test.go
    module_register_example.go
```

## 已生成模块

- `DBModule`：包含 `BeforeLoad/Load/AfterLoad/Close`
- 显式依赖声明：`Depends() []string`
- `Close()` 幂等
- `Load(cfg interface{})` 强类型校验与可操作错误

## 测试覆盖

- 生命周期顺序
- 依赖顺序稳定性
- required 模块缺失配置错误
- Close 幂等
- Load 非法配置类型错误

