import type { Player } from './api'

export interface SessionMatch {
  id: number; round: number; player_a_id: number; player_b_id: number
  player_a_name: string; player_b_name: string
  score_a: number; score_b: number
  rating_change_a: number; rating_change_b: number
  winner_id: number; played: boolean; forfeit: boolean
}

export interface SessionPlayer {
  id: number; name: string; starting_rating: number
  wins: number; losses: number
}

export interface SessionDetail {
  id: number; name: string; status: string; created_at: string
  players: SessionPlayer[]; matches: SessionMatch[]
}

export function unplayedCount(s: SessionDetail) {
  return s.matches?.filter(m => !m.played).length || 0
}

export function sessionChange(matches: SessionMatch[], pid: number): number {
  let total = 0
  for (const m of matches) {
    if (!m.played) continue
    if (m.player_a_id === pid) total += m.rating_change_a
    if (m.player_b_id === pid) total += m.rating_change_b
  }
  return total
}

export function sessionDisplayRating(p: SessionPlayer, matches: SessionMatch[]): number {
  return p.starting_rating + sessionChange(matches, p.id)
}

export function changeSign(v: number): string { return v >= 0 ? '+' : '' }

export function selectedPlayers(players: Player[], selectedIDs: Set<number>): Player[] {
  return Array.from(selectedIDs).map(id => players.find(p => p.id === id)!).filter(Boolean)
}
