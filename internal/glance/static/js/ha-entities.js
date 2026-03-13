const STATE_ON = 'on';
const TOGGLEABLE_DOMAINS = new Set(['switch', 'light', 'input_boolean']);

export default async function (element) {
    const widgetId = element.dataset.widgetId;
    const baseURL = (typeof pageData !== 'undefined' && pageData.baseURL) ? pageData.baseURL : '';

    async function request(method, suffix, body) {
        const url = `${baseURL}/api/widgets/ha-entities/${widgetId}${suffix}`;
        const opts = { method };
        if (body !== undefined) {
            opts.headers = { 'Content-Type': 'application/json' };
            opts.body = JSON.stringify(body);
        }
        const resp = await fetch(url, opts);
        if (!resp.ok) {
            const msg = await resp.text().catch(() => resp.statusText);
            throw new Error(`Request failed (${resp.status}): ${msg}`);
        }
        return resp.json();
    }

    // Collect all tile elements indexed by entity_id
    const tiles = {};
    for (const tile of element.querySelectorAll('.ha-entities-tile')) {
        const entityId = tile.dataset.entityId;
        if (entityId) tiles[entityId] = tile;
    }

    function applyState(tile, state) {
        const domain = tile.dataset.domain;
        const isToggleable = tile.dataset.toggleable === 'true';
        const isSensor = tile.dataset.sensor === 'true';
        const valueEl = tile.querySelector('.ha-entities-value');

        tile.classList.remove('is-on', 'is-off', 'is-unknown', 'is-loading');

        if (!state) {
            tile.classList.add('is-unknown');
            if (valueEl) valueEl.textContent = '—';
            return;
        }

        const stateStr = state.state;

        if (isToggleable || domain === 'binary_sensor') {
            tile.classList.add(stateStr === STATE_ON ? 'is-on' : 'is-off');
        } else if (isSensor) {
            // Sensor: always show as active, display value
            tile.classList.add('is-on');
            if (valueEl) {
                valueEl.textContent = state.unit ? `${stateStr} ${state.unit}` : stateStr;
            }
        } else {
            tile.classList.add('is-on');
        }
    }

    function applyStates(statesMap) {
        for (const [entityId, tile] of Object.entries(tiles)) {
            applyState(tile, statesMap[entityId] || null);
        }
    }

    // Attach click handlers for toggleable tiles
    for (const tile of Object.values(tiles)) {
        if (tile.dataset.toggleable !== 'true') continue;

        tile.addEventListener('click', async () => {
            if (tile.classList.contains('is-pending')) return;

            tile.classList.add('is-pending');
            try {
                const updated = await request('POST', '/toggle', { entity_id: tile.dataset.entityId });
                applyStates(updated);
            } catch (err) {
                console.error('ha-entities toggle failed:', err);
            } finally {
                tile.classList.remove('is-pending');
            }
        });
    }

    // Attach click handlers for script tiles
    for (const tile of element.querySelectorAll('.ha-entities-tile.is-script')) {
        tile.addEventListener('click', async () => {
            if (tile.classList.contains('is-pending')) return;

            tile.classList.add('is-pending');
            try {
                await request('POST', '/run', { entity_id: tile.dataset.entityId });
                tile.classList.add('is-executed');
                setTimeout(() => tile.classList.remove('is-executed'), 2000);
            } catch (err) {
                console.error('ha-entities run failed:', err);
            } finally {
                tile.classList.remove('is-pending');
            }
        });
    }

    // Mark all tiles as loading
    for (const tile of Object.values(tiles)) {
        tile.classList.add('is-loading');
    }

    // Initial state load
    try {
        const states = await request('GET', '');
        applyStates(states);
    } catch (err) {
        for (const tile of Object.values(tiles)) {
            tile.classList.remove('is-loading');
            tile.classList.add('is-unknown');
            const valueEl = tile.querySelector('.ha-entities-value');
            if (valueEl) valueEl.textContent = 'Fout';
        }
        console.error('ha-entities load failed:', err);
    }
}
