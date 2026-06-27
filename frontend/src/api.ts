const BASE = '/api'

export interface Player {
  id: number
  name: string
  initial_rating: number
  current_rating: number
  created_at: string
}

export interface RankingEntry extends Player {
  matches_played: number
  wins: number
  losses: number
  forfeit_wins: number
  forfeits: number
  win_rate: number
}

export interface Match {
  id: number
  player_a_id: number
  player_b_id: number
  player_a_name: string
  player_b_name: string
  score_a: number
  score_b: number
  rating_change_a: number
  rating_change_b: number
  winner_id: number
  forfeit: boolean
  played_at: string
}

export interface PlayerDetail {
  player: Player
  matches: Match[]
  wins: number
  losses: number
  forfeit_wins: number
  forfeits: number
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || res.statusText)
  }
  return res.json()
}

export const api = {
  getRankings: () => request<RankingEntry[]>('/rankings'),
  getPlayers: () => request<Player[]>('/players'),
  getPlayer: (id: number) => request<PlayerDetail>('/players/' + id),
  createPlayer: (data: { name: string; initial_rating?: number }) =>
    request<Player>('/players', { method: 'POST', body: JSON.stringify(data) }),
  getMatches: () => request<Match[]>('/matches'),
  createMatch: (data: { player_a_id: number; player_b_id: number; score_a: number; score_b: number }) =>
    request<Match>('/matches', { method: 'POST', body: JSON.stringify(data) }),
}
