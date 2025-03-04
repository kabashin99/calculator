const API_URL = 'http://localhost:8080/api/v1';

// Отправка выражения
document.getElementById('expressionForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const expression = document.getElementById('expressionInput').value;

    try {
        const response = await fetch(`${API_URL}/calculate`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ expression })
        });

        if (!response.ok) throw new Error('Ошибка отправки');
        document.getElementById('expressionInput').value = '';
        loadExpressions(); // Обновить список
    } catch (error) {
        alert(error.message);
    }
});

// Загрузка списка выражений
async function loadExpressions() {
    try {
        const response = await fetch(`${API_URL}/expressions`);
        const data = await response.json();
        renderExpressions(data.expressions);
    } catch (error) {
        console.error('Ошибка загрузки:', error);
    }
}

// Отрисовка выражений
function renderExpressions(expressions) {
    const list = document.getElementById('expressionList');
    list.innerHTML = '<h2>История вычислений</h2>';

    expressions.forEach(expr => {
        const div = document.createElement('div');
        div.className = `expression-item status-${expr.status}`;
        div.innerHTML = `
            ID: ${expr.id}<br>
            Выражение: ${expr.expression || 'N/A'}<br>
            Статус: ${expr.status}<br>
            Результат: ${expr.result || '—'}
        `;
        list.appendChild(div);
    });
}

// Автообновление каждые 3 секунды
setInterval(loadExpressions, 3000);
loadExpressions(); // Первоначальная загрузка