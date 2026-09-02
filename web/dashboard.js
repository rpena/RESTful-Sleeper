import { dashboardStyles, escapeHTML, formatNumber, matchupStatus } from "./sleeper-dashboard-core.js";

const form = document.querySelector("#settings");
const endpoint = document.querySelector("#endpoint");
const content = document.querySelector("#content");
const lastUpdated = document.querySelector("#last-updated");

function render(data) {
  const matchups = Array.isArray(data.matchups) ? data.matchups : [];
  const standings = Array.isArray(data.standings) ? data.standings : [];
  const waivers = Array.isArray(data.waiver_pickups) ? data.waiver_pickups : [];
  const yourTeam = data.matchup?.your_team;
  const featured = yourTeam ? matchups.find((item) => item.teams?.some((team) => team.roster_id === yourTeam.roster_id)) : null;
  const others = featured ? matchups.filter((item) => item !== featured) : matchups;
  const league = data.league || {};

  content.innerHTML = `
    <section class="section featured-section">
      <div class="section-heading"><h2>${featured ? "Your matchup" : "Gameday status"}</h2><span>${featured ? `matchup ${escapeHTML(featured.matchup_id)}` : escapeHTML(data.status || "")}</span></div>
      ${featured ? `<div class="featured-matchup">${matchup(featured)}</div>` : `<div class="empty"><strong>${data.game_day === false ? "No NFL games scheduled today" : "No matchup data yet"}</strong><span>${escapeHTML(data.status || "Check the API response and try again.")}</span></div>`}
    </section>
    ${others.length ? `<section class="section secondary"><div class="section-heading"><h2>Other matchups</h2><span>${others.length} games</span></div><div class="matchups">${others.map(matchup).join("")}</div></section>` : ""}
    ${waivers.length ? `<section class="section secondary"><div class="section-heading"><h2>Waiver watch</h2><span>top adds</span></div><div class="waivers">${waivers.slice(0, 5).map(waiver).join("")}</div></section>` : ""}
    ${standings.length ? `<section class="section secondary standings-section"><div class="section-heading"><h2>Leaderboard</h2><span>points for</span></div><div class="table">${standings.slice(0, 10).map(standing).join("")}</div></section>` : ""}
  `;
  document.querySelector("#league-name").textContent = league.name || "Sleeper league";
  document.querySelector("#week-label").textContent = `Week ${data.week ?? "-"} / ${league.season || ""}`;
}

function matchup(item, index) {
  const teams = Array.isArray(item.teams) ? item.teams : [];
  return `<article class="matchup"><div class="matchup-label"><span>Matchup ${escapeHTML(item.matchup_id ?? index + 1)}</span><span class="status">${matchupStatus(teams)}</span></div>${teams.map(team).join("")}</article>`;
}

function team(item) {
  const current = Number(item.current_points) || 0;
  const projected = Number(item.projected_points) || 0;
  const max = Math.max(current, projected, 1);
  return `<div class="team"><div class="team-top"><span class="team-name">${escapeHTML(item.display_name || `Roster ${item.roster_id}`)}</span><strong>${formatNumber(current)}</strong></div><div class="bar"><i style="width:${Math.min(100, current / max * 100)}%"></i></div><div class="team-meta"><span>${item.players_completed || 0} played <b>·</b> ${item.players_remaining || 0} remaining</span><span>${item.projection_available ? `proj. ${formatNumber(projected)}` : "projection pending"}</span></div></div>`;
}

function standing(item) { return `<div class="standing"><span class="rank">${escapeHTML(item.rank ?? "-")}</span><span class="standing-name">${escapeHTML(item.display_name || `Roster ${item.roster_id}`)}</span><span class="record">${item.wins ?? 0}-${item.losses ?? 0}-${item.ties ?? 0}</span><strong>${formatNumber(item.points_for)}</strong></div>`; }
function waiver(item) { return `<div class="waiver"><span class="waiver-badge">${escapeHTML(item.position || "FA")}</span><span class="waiver-name">${escapeHTML(item.name || item.player_id)}</span><strong>+${escapeHTML(item.adds ?? 0)}</strong></div>`; }

async function load() {
  content.innerHTML = `<div class="empty"><strong>Loading dashboard</strong><span>Fetching the latest cached data.</span></div>`;
  try {
    const response = await fetch(endpoint.value);
    if (!response.ok) throw new Error(`API returned ${response.status}`);
    render(await response.json());
    lastUpdated.textContent = `Updated ${new Date().toLocaleTimeString()}`;
  } catch (error) {
    content.innerHTML = `<div class="empty"><strong>Unable to load dashboard</strong><span>${escapeHTML(error.message)}</span></div>`;
  }
}

form.addEventListener("submit", (event) => { event.preventDefault(); load(); });
document.querySelector("#refresh").addEventListener("click", load);
document.head.append(Object.assign(document.createElement("style"), { textContent: dashboardStyles() }));
load();
