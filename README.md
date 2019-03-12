# v2ray-sspanel-v3-mod_Uim-plugin
# 免费版本
# 使用教程请看wiki
### 收费版 [请点击这里](https://github.com/rico93/pay-v2ray-sspanel-v3-mod_Uim-plugin/)
## 公告

## Thanks
1. 感恩的 [ColetteContreras's repo](https://github.com/ColetteContreras/v2ray-ssrpanel-plugin). 让我一个go小白有了下手地。主要起始框架来源于这里
2. 感恩 [eycorsican](https://github.com/eycorsican) 在v2ray-core [issue](https://github.com/v2ray/v2ray-core/issues/1514), 促成了go版本提上日程


# 划重点
1. 目前端口设置为0，才会监听本地，不再是443
2. 已经适配了中转，必须用我自己维护的[panel](https://github.com/rico93/ss-panel-v3-mod_Uim)
   
## 项目状态

支持 [ss-panel-v3-mod_Uim](https://github.com/NimaQu/ss-panel-v3-mod_Uim) 的 webapi。 目前自己也尝试维护了一个版本, [panel](https://github.com/rico93/ss-panel-v3-mod_Uim)

目前只适配了流量记录、服务器是否在线、在线人数,在线ip上报、负载、中转，后端根据前端的设定自动调用 API 增加用户。

v2ray 后端 kcp、tcp、ws 都是多用户共用一个端口。

也可作为 ss 后端一个用户一个端口。


