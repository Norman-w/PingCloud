export interface Bulletin {
  id: number
  date: string
  title: string
  content: string
}

export const bulletins: Bulletin[] = [
  {
    id: 2,
    date: '2026-06-29',
    title: '积分系统重大修复：全员积分已重新核算',
    content: `发现一个bug：老活动的积分在结算时被重复加了一遍，导致部分球员积分虚高（如陆瑜显示1760实际应为1561）。

**已修复：**
- 全员积分按「初始分 + 每场累计变化」重新核算
- 后续新活动不会再有此问题（已加防重复结算机制）

**主要变化：**
- 李博文 1606→1625 · 陆瑜 1760→1561 · 王帅 1589→1487
- 姜舒航 1500→1528 · 赵慧明 1481→1507
- 张志远 1452 不变 · 姜堡垒 1452 不变 · 崔盈龙 1404 不变

积分数据已备份，如有疑问随时沟通。`
  },
  {
    id: 1,
    date: '2026-06-29',
    title: '6.28比赛数据异常修复说明',
    content: `昨晚比赛数据出现了一个bug：**姜堡垒**中场加入后，系统没记录到他入场的积分，导致计分偏高。

今早已修复，所有场次按正确积分重新计算，排名已恢复正常。

受影响的球员：
- 姜堡垒 1500 → 1452（-48）
- 陆瑜 +40，王帅 +16，李博文 +8，崔盈龙 -16
- 其余人不受影响

给大家带来困扰抱歉 🙏 有事随时说。`
  }
]

const STORAGE_KEY = 'pingpong_read_bulletins'

export function getReadIds(): number[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : []
  } catch { return [] }
}

export function markRead(id: number) {
  const ids = getReadIds()
  if (!ids.includes(id)) {
    ids.push(id)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(ids))
  }
  return ids
}

export function unreadCount(): number {
  return bulletins.filter(b => !getReadIds().includes(b.id)).length
}
