const trashIconSvg = `<svg fill="currentColor" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16">
  <path fill-rule="evenodd" d="M5 3.25V4H2.75a.75.75 0 0 0 0 1.5h.3l.815 8.15A1.5 1.5 0 0 0 5.357 15h5.285a1.5 1.5 0 0 0 1.493-1.35l.815-8.15h.3a.75.75 0 0 0 0-1.5H11v-.75A2.25 2.25 0 0 0 8.75 1h-1.5A2.25 2.25 0 0 0 5 3.25Zm2.25-.75a.75.75 0 0 0-.75.75V4h3v-.75a.75.75 0 0 0-.75-.75h-1.5ZM6.05 6a.75.75 0 0 1 .787.713l.275 5.5a.75.75 0 0 1-1.498.075l-.275-5.5A.75.75 0 0 1 6.05 6Zm3.9 0a.75.75 0 0 1 .712.787l-.275 5.5a.75.75 0 0 1-1.498-.075l.275-5.5a.75.75 0 0 1 .786-.711Z" clip-rule="evenodd" />
</svg>`;

function formatDue(due) {
    if (!due) return '';
    // HA returns "2026-03-15" (date) or "2026-03-15T10:00:00" (datetime)
    const d = new Date(due.includes('T') ? due : due + 'T00:00:00');
    if (isNaN(d.getTime())) return due;
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
}

export default async function (element) {
    const widgetId = element.dataset.widgetId;
    const emptyText = element.dataset.emptyText || 'No open tasks!';
    const addPlaceholder = element.dataset.addPlaceholder || 'Add a task\u2026';
    const baseURL = (typeof pageData !== 'undefined' && pageData.baseURL) ? pageData.baseURL : '';

    async function request(method, suffix, body) {
        const url = `${baseURL}/api/widgets/ha-todo/${widgetId}${suffix}`;
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

    function showError(message) {
        const err = document.createElement('p');
        err.className = 'ha-todo-error';
        err.textContent = '\u26a0 ' + message;
        element.prepend(err);
        setTimeout(() => err.remove(), 5000);
    }

    function makeItem(item) {
        const li = document.createElement('li');
        li.className = 'ha-todo-item';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'ha-todo-checkbox';
        checkbox.checked = item.status === 'completed';

        checkbox.addEventListener('change', async () => {
            checkbox.disabled = true;
            const newStatus = checkbox.checked ? 'completed' : 'needs_action';
            // uid may be empty for some HA integrations — fall back to summary
            const uid = item.uid || item.summary;
            try {
                const newItems = await request('PATCH', '/items', { uid, status: newStatus });
                renderItems(newItems || []);
            } catch (err) {
                checkbox.checked = !checkbox.checked;
                checkbox.disabled = false;
                showError(err.message);
            }
        });

        const body = document.createElement('div');
        body.className = 'ha-todo-body';

        const text = document.createElement('span');
        text.className = 'ha-todo-text' + (item.status === 'completed' ? ' completed' : '');
        text.textContent = item.summary;
        body.appendChild(text);

        if (item.due) {
            const due = document.createElement('span');
            due.className = 'ha-todo-due';
            due.textContent = formatDue(item.due);
            body.appendChild(due);
        }

        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'ha-todo-delete';
        deleteBtn.innerHTML = trashIconSvg;
        deleteBtn.title = 'Delete task';

        deleteBtn.addEventListener('click', async () => {
            deleteBtn.disabled = true;
            const uid = item.uid || item.summary;
            try {
                const newItems = await request('DELETE', '/items', { uid });
                renderItems(newItems || []);
            } catch (err) {
                deleteBtn.disabled = false;
                showError(err.message);
            }
        });

        li.appendChild(checkbox);
        li.appendChild(body);
        li.appendChild(deleteBtn);
        return li;
    }

    function renderItems(items) {
        const activeItems = items.filter(i => i.status === 'needs_action');
        const doneItems = items.filter(i => i.status === 'completed');

        element.innerHTML = '';

        if (activeItems.length === 0 && doneItems.length === 0) {
            const empty = document.createElement('p');
            empty.className = 'ha-todo-empty';
            empty.textContent = emptyText;
            element.appendChild(empty);
        } else {
            const list = document.createElement('ul');
            list.className = 'ha-todo-list';
            for (const item of activeItems) {
                list.appendChild(makeItem(item));
            }
            element.appendChild(list);
        }

        if (doneItems.length > 0) {
            const label = document.createElement('p');
            label.className = 'ha-todo-section-label';
            label.textContent = 'Completed';
            element.appendChild(label);

            const doneList = document.createElement('ul');
            doneList.className = 'ha-todo-list';
            for (const item of doneItems) {
                doneList.appendChild(makeItem(item));
            }
            element.appendChild(doneList);
        }

        // Add-item input row
        const inputRow = document.createElement('div');
        inputRow.className = 'ha-todo-input';

        const plusIcon = document.createElement('span');
        plusIcon.className = 'ha-todo-plus-icon';

        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'ha-todo-input-field';
        input.placeholder = addPlaceholder;
        input.setAttribute('spellcheck', 'false');

        input.addEventListener('keydown', async (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                const summary = input.value.trim();
                if (!summary) return;
                input.disabled = true;
                try {
                    const newItems = await request('POST', '/items', { summary });
                    renderItems(newItems || []);
                } catch (err) {
                    input.disabled = false;
                    showError(err.message);
                }
            } else if (e.key === 'Escape') {
                input.value = '';
                input.blur();
            }
        });

        plusIcon.addEventListener('click', () => input.focus());
        inputRow.appendChild(plusIcon);
        inputRow.appendChild(input);
        element.appendChild(inputRow);
    }

    // Initial render: show loading state then fetch
    const loading = document.createElement('p');
    loading.className = 'ha-todo-loading';
    loading.textContent = 'Loading\u2026';
    element.appendChild(loading);

    try {
        const items = await request('GET', '');
        renderItems(items || []);
    } catch (err) {
        element.innerHTML = '';
        const errElem = document.createElement('p');
        errElem.className = 'ha-todo-error';
        errElem.textContent = '\u26a0 Failed to load: ' + err.message;
        element.appendChild(errElem);
    }
}
