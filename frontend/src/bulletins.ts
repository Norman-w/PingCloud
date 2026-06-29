export interface Bulletin {
  id: number
  date: string
  title: string
  content: string
}

export const bulletins: Bulletin[] = [
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
