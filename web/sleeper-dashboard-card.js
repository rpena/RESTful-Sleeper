class SleeperDashboardCard extends HTMLElement {
  static getStubConfig() {
    return { entity: "sensor.sleeper_dashboard" };
  }

  setConfig(config) {
    if (!config.entity) {
      throw new Error("Sleeper Dashboard Card requires an entity");
    }
    this.config = {
      title: "Gameday board",
      entity: config.entity,
      show_standings: true,
      show_waivers: true,
      show_raw: false,
      ...config,
    };
    if (!this.shadowRoot) {
      this.attachShadow({ mode: "open" });
    }
    this.render();
  }

  set hass(hass) {
    this._hass = hass;
    this.render();
  }

  getCardSize() {
    return 12;
  }

  render() {
    if (!this.config || !this.shadowRoot) return;
    const state = this._hass?.states?.[this.config.entity];
    const data = state?.attributes || {};
    const matchups = Array.isArray(data.matchups) ? data.matchups : [];
    const standings = Array.isArray(data.standings) ? data.standings : [];
    const waivers = Array.isArray(data.waiver_pickups) ? data.waiver_pickups : [];
    const league = data.league || {};
    const yourTeam = data.matchup?.your_team;
    const featuredMatchup = yourTeam
      ? matchups.find((item) => item.teams?.some((team) => team.roster_id === yourTeam.roster_id))
      : null;
    const otherMatchups = featuredMatchup
      ? matchups.filter((item) => item !== featuredMatchup)
      : matchups;

    this.shadowRoot.innerHTML = `
      <style>${this.styles()}</style>
      <ha-card>
        <div class="shell">
          <header class="header">
            <div>
              <div class="eyebrow">${this.escape(league.name || "Sleeper league")}</div>
              <h1>${this.escape(this.config.title)}</h1>
              <div class="subhead">Week ${this.escape(data.week ?? "-")} <span class="dot">•</span> ${this.escape(league.season || "")}</div>
            </div>
            <button class="refresh" title="Refresh dashboard data" aria-label="Refresh dashboard data">↻</button>
          </header>
          ${featuredMatchup ? `<section class="section featured-section"><div class="section-heading"><h2>Your matchup</h2><span>matchup ${this.escape(featuredMatchup.matchup_id)}</span></div><div class="featured-matchup">${this.matchup(featuredMatchup, 0)}</div></section>` : this.emptyState(state)}
          ${otherMatchups.length ? `<section class="section secondary"><div class="section-heading"><h2>Other matchups</h2><span>${otherMatchups.length} games</span></div><div class="matchups">${otherMatchups.map((item, index) => this.matchup(item, index)).join("")}</div></section>` : ""}
          ${this.config.show_waivers && waivers.length ? `<section class="section secondary"><div class="section-heading"><h2>Waiver watch</h2><span>top adds</span></div><div class="waivers">${waivers.slice(0, 5).map((player) => this.waiver(player)).join("")}</div></section>` : ""}
          ${this.config.show_standings && standings.length ? `<section class="section secondary standings-section"><div class="section-heading"><h2>Leaderboard</h2><span>points for</span></div><div class="table">${standings.slice(0, 10).map((team) => this.standing(team)).join("")}</div></section>` : ""}
          ${this.config.show_raw ? `<details class="raw"><summary>Raw response</summary><pre>${this.escape(JSON.stringify(data, null, 2))}</pre></details>` : ""}
        </div>
      </ha-card>
    `;

    this.shadowRoot.querySelector(".refresh")?.addEventListener("click", () => {
      this._hass?.callService("homeassistant", "update_entity", { entity_id: this.config.entity });
    });
  }

  matchup(item, index) {
    const teams = Array.isArray(item.teams) ? item.teams : [];
    return `<article class="matchup">
      <div class="matchup-label"><span>Matchup ${this.escape(item.matchup_id ?? index + 1)}</span><span class="status">${this.status(teams)}</span></div>
      <div class="teams">${teams.map((team) => this.team(team)).join("")}</div>
    </article>`;
  }

  team(team) {
    const current = Number(team.current_points) || 0;
    const projected = Number(team.projected_points) || 0;
    const max = Math.max(current, projected, 1);
    const remaining = Number(team.players_remaining) || 0;
    const completed = Number(team.players_completed) || 0;
    return `<div class="team">
      <div class="team-top"><span class="team-name">${this.escape(team.display_name || `Roster ${team.roster_id}`)}</span><strong>${this.number(current)}</strong></div>
      <div class="bar"><i style="width:${Math.min(100, current / max * 100)}%"></i></div>
      <div class="team-meta"><span>${completed} played <b>·</b> ${remaining} remaining</span><span>${team.projection_available ? `proj. ${this.number(projected)}` : "projection pending"}</span></div>
    </div>`;
  }

  standing(team) {
    return `<div class="standing"><span class="rank">${this.escape(team.rank ?? "-")}</span><span class="standing-name">${this.escape(team.display_name || `Roster ${team.roster_id}`)}</span><span class="record">${team.wins ?? 0}-${team.losses ?? 0}-${team.ties ?? 0}</span><strong>${this.number(team.points_for)}</strong></div>`;
  }

  waiver(player) {
    return `<div class="waiver"><span class="waiver-badge">${this.escape(player.position || "FA")}</span><span class="waiver-name">${this.escape(player.name || player.player_id)}</span><strong>+${this.escape(player.adds ?? 0)}</strong></div>`;
  }

  status(teams) {
    const remaining = teams.reduce((sum, team) => sum + (Number(team.players_remaining) || 0), 0);
    return remaining ? `${remaining} players left` : "complete";
  }

  emptyState(state) {
    if (!state) return `<div class="empty"><strong>Waiting for dashboard data</strong><span>Check the REST sensor entity in Home Assistant.</span></div>`;
	if (state.attributes?.game_day === false) return `<div class="empty"><strong>No NFL games scheduled today</strong><span>${this.escape(state.attributes.status || "Dashboard refresh is paused until the next game window.")}</span></div>`;
    return `<div class="empty"><strong>No matchup data yet</strong><span>${this.escape(state.state || "The selected week has no data.")}</span></div>`;
  }

  number(value) {
    return Number(value || 0).toFixed(2);
  }

  escape(value) {
    return String(value ?? "").replace(/[&<>"']/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character]);
  }

  styles() {
    return `
      :host { --ink: #15232b; --muted: #718087; --line: #dce4e5; --accent: #d85d3f; --mint: #dcefe9; --paper: #f7faf8; display: block; }
      * { box-sizing: border-box; }
      ha-card { overflow: hidden; color: var(--ink); background: var(--paper); border: 1px solid var(--line); border-radius: 14px; box-shadow: none; }
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
}

customElements.define("sleeper-dashboard-card", SleeperDashboardCard);
window.customCards = window.customCards || [];
window.customCards.push({
  type: "sleeper-dashboard-card",
  name: "Sleeper Gameday Dashboard",
  description: "League-wide Sleeper matchup board with scores, projections, standings, and waivers.",
  preview: true,
});
