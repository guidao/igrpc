# 交互式grpc命令行工具

主要用于通过命令行的方式调用grpc服务. 相似于grpcurl.

# feature

[X] 列表service
[X] 列出method
[X] 描述method
[X] 描述参数定义
[X] 调用grpc函数
[ ] 支持补全
[ ] 支持历史记录


## 启动

```
$ igrpc localhost:8080
igrpc> list
grpc.reflection.v1alpha.ServerReflection
pb.Admin
pb.App
igrpc> list pb.Admin
pb.Admin.AppList
pb.Admin.CheckLogin
pb.Admin.CreateApp
pb.Admin.CreateOrg
pb.Admin.OrgList
pb.Admin.UpdateApp
pb.Admin.UpdateOrg
igrpc> desc pb.Admin.AppList
pb.Admin.AppList rpc AppList ( .pb.AppListRequest ) returns ( .pb.AppListResponse );
igrpc> call pb.Admin.AppList {}

```

## list 命令

如果不带参数，list命令列出远端的service，如果带service参数，则列出service的函数

## desc 命令

用户查看service支持的函数跟参数的定义。

## call 命令

调用函数,目前仅支持json格式


