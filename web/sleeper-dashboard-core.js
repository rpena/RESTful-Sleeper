export function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character]);
}

export function formatNumber(value) {
  return Number(value || 0).toFixed(2);
}

export function matchupStatus(teams) {
  const remaining = teams.reduce((sum, team) => sum + (Number(team.players_remaining) || 0), 0);
  return remaining ? `${remaining} players left` : "complete";
}

export function dashboardStyles() {
  return `
    :host { --ink: #15232b; --muted: #718087; --line: #dce4e5; --accent: #d85d3f; --paper: #f7faf8; display: block; }
    * { box-sizing: border-box; }
    ha-card, .dashboard-card { overflow: hidden; color: var(--ink); background: var(--paper); border: 1px solid var(--line); border-radius: 14px; box-shadow: none; }
    .shell { padding: 22px; }
    .header { display: flex; align-items: flex-start; justify-content: space-between; padding-bottom: 20px; border-bottom: 1px solid var(--line); }
    .eyebrow, .section-heading span, .subhead, .team-meta, .status, .record { color: var(--muted); font-size: 11px; letter-spacing: .08em; text-transform: uppercase; }
    h1, h2 { margin: 0; font-family: Georgia, serif; font-weight: 500; letter-spacing: 0; }
    h1 { font-size: 30px; line-height: 1.05; margin-top: 5px; }
    h2 { font-size: 19px; }
    .subhead { margin-top: 9px; letter-spacing: .03em; text-transform: none; }
    .dot { color: var(--accent); padding: 0 5px; }
    button { font: inherit; }
    .refresh { width: 38px; height: 38px; border: 1px solid var(--line); border-radius: 50%; background: white; color: var(--accent); font-size: 24px; line-height: 1; cursor: pointer; }
    .section { padding-top: 21px; }
    .section.secondary { border-top: 1px solid var(--line); margin-top: 24px; }
    .section-heading { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 11px; }
    .matchups { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
    .featured-matchup { max-width: 760px; }
    .featured-matchup .matchup { border: 2px solid var(--accent); padding: 18px; }
    .featured-matchup .team-top strong { font-size: 31px; }
    .standings-section .table { max-width: 680px; }
    .matchup { background: white; border: 1px solid var(--line); border-radius: 9px; padding: 14px; }
    .matchup-label { display: flex; justify-content: space-between; align-items: center; color: var(--muted); font-size: 11px; letter-spacing: .07em; text-transform: uppercase; margin-bottom: 13px; }
    .status { color: var(--accent); font-size: 10px; letter-spacing: .04em; }
    .team + .team { border-top: 1px solid var(--line); margin-top: 14px; padding-top: 14px; }
    .team-top { display: flex; justify-content: space-between; gap: 12px; align-items: baseline; }
    .team-name { font-weight: 650; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .team-top strong { font-family: Georgia, serif; font-size: 25px; font-weight: 500; white-space: nowrap; }
    .bar { height: 5px; background: #e9eeee; margin-top: 8px; overflow: hidden; border-radius: 5px; }
    .bar i { display: block; height: 100%; background: var(--accent); border-radius: 5px; transition: width .5s ease; }
    .team-meta { display: flex; justify-content: space-between; gap: 10px; margin-top: 8px; font-size: 10px; letter-spacing: .02em; text-transform: none; }
    .team-meta b { color: var(--accent); padding: 0 3px; }
    .table, .waivers { background: white; border: 1px solid var(--line); border-radius: 9px; overflow: hidden; }
    .standing, .waiver { display: grid; align-items: center; min-height: 42px; padding: 0 13px; border-bottom: 1px solid var(--line); gap: 10px; }
    .standing:last-child, .waiver:last-child { border-bottom: 0; }
    .standing { grid-template-columns: 22px 1fr auto 62px; }
    .rank { color: var(--accent); font-family: Georgia, serif; font-size: 17px; }
    .standing-name, .waiver-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .standing strong, .waiver strong { text-align: right; }
    .record { font-size: 10px; letter-spacing: .02em; text-transform: none; }
    .waiver { grid-template-columns: 35px 1fr auto; }
    .waiver-badge { color: var(--accent); font-size: 10px; font-weight: 700; letter-spacing: .04em; }
    .empty { padding: 30px 4px 18px; display: grid; gap: 6px; color: var(--muted); }
    .empty strong { color: var(--ink); font-family: Georgia, serif; font-size: 19px; font-weight: 500; }
    .raw { margin-top: 22px; border-top: 1px solid var(--line); padding-top: 14px; color: var(--muted); font-size: 12px; }
    .raw summary { cursor: pointer; }
    pre { max-height: 360px; overflow: auto; padding: 12px; background: #17252b; color: #dcefe9; border-radius: 8px; font-size: 11px; white-space: pre-wrap; }
    @media (max-width: 500px) { .shell { padding: 16px; } h1 { font-size: 26px; } .matchups { grid-template-columns: 1fr; } .team-meta { font-size: 9px; } }
  `;
}
