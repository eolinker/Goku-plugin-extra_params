### 插件名称

| 类别 |  名称 |  字段  | 属性  |
| ------------ | ------------ | ------------ |------------ |
| API插件 | 额外参数 | goku-extra_params  | 安全防御 |


### 功能描述

开启该插件后，不需要用户传某些参数值，网关会在转发时自动带上这些参数，支持header、body、query参数。
额外参数仅支持 **表单** 类型与 **json** 类型：
* formdata的参数值须为string类型，头部补充Conent-Type:x-www-form-urlencoded。
* 若额外参数是json类型，需在头部补充Content-Type:application/json。
* 参数类型为表单时支持同名参数。


### 配置页面

进入控制台 >> 策略管理 >> 某策略 >> API插件 >> 额外参数插件：


![](http://data.eolinker.com/course/v6x1ZXl19cf9a61e29c11c04ad602f865135e58ba663c2b)

### 配置参数

| 参数名 | 说明   | 
| ------------ | ------------ |  
|  params |额外参数列表 | 
| paramName  | 参数名 |
| paramPosition  | 参数位置 |  
| paramValue  | 参数值 | 
| paramConflictSolution  |  参数冲突时的处理方式 [origin/convert/error] |

参数冲突说明：
额外参数插件配置了参数A的值，但是直接请求时也传了参数A，此时为参数出现冲突，参数A实际上会接收两个参数值。
* convert：参数出现冲突时，取映射后的参数，即配置的值
* origin：参数出现冲突时，取映射前的参数，即实际传的值
* error：请求时报错，"param_name"has a conflict.

若paramConflictSolution为空，视为使用默认值convert。

### 配置示例

```
{
    "params": [
        {
            "paramName": "test",
            "paramPosition": "header",
            "paramValue": "extra_param",
            "paramConflictSolution":"convert"
        }
    ]
}
```

### 请求参数

| 参数名 | 说明  | 必填  |   值可能性   |  参数位置 |
| :----------- | :----------- | :----------- | :----------- | :----------- |
|  Strategy-Id | 策略ID  | 是 |   |  header  | 
|  Content-Type | 数据类型  | 是 | x-www-form-urlencoded 或 application/json   |  header  | 

若该 test 参数为表单参数，则请求头部填写 Conent-Type:x-www-form-urlencoded。
若该 test 参数为Json参数，则请求头部需加 Conent-Type:application/json。
