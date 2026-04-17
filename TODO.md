- [x] ~~设备同步通道时未同步到的通道应该逻辑删除,以免保留垃圾数据~~ ✅ **已完成** - 使用 deltime 实现软删除，Catalog 同步完成时清理旧通道

---

- [x] ~~通道停止流的时候用的是流ID: deviceid_channelid，需要统一流ID命名规范~~ ✅ **已完成** - 统一格式：live_{device}_{channel}、replay_{device}_{channel}_{time}、talk_{device}_{channel}
- [ ] 根据 review.md 修复代码问题（P0: 并发安全、竞态条件）
- [ ] 设备和通道添加经纬度字段用于记录最新的经纬度
- [ ] 使用 redis 记录经纬度用于查询和订阅经纬度（要求30分钟内没有上传则认为是新的一段经纬度轨迹）