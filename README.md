utils/tools.go中包含了开发中常用的方法，main.go中有实例调用方法；
可修改/proto下的协议文件内容，运行脚本./pbgen.sh，生成的pb.go会保存到/pbs覆盖原文件；
db以mongo为例，采用gogoproto tag来对应bson字段。
