const EXAMPLES = {
    pow: `pow(m, n);
init: result := 1;
      goto test;
test: if n < 1 goto end else loop;
loop: result := result * m;
      n := n - 1;
      goto test;
end: return result;`,
    ackermann: `ackerman(m, n):
ack: if m = 0 goto done else next;
next: if n = 0 goto ack0 else ack1;
done: return n + 1;
ack0: n := 1;
  goto ack2;
ack1: n := n - 1;
  n := call ack;
  goto ack2;
ack2: m := m - 1;
  n := call ack;
  return n;`,
    turing: `TuringMachine(Q, Right):
init: Qtail := Q;
      Left := '();
      goto loop;
loop: if Qtail = '() goto stop else cont;
cont: Instruction := hd(Qtail);
      Qtail := tl(Qtail);
      Operator := hd(tl(Instruction));
      if Operator = 'right goto do_right else cont1;
cont1: if Operator = 'left goto do_left else cont2;
cont2: if Operator = 'write goto do_write else cont3;
cont3: if Operator = 'goto goto do_goto else cont4;
cont4: if Operator = 'if goto do_if else error;

do_right: Left := cons(hd(Right), Left);
          Right := tl(Right);
          goto loop;
do_left:  Right := cons(hd(Left), Right);
          Left := tl(Left);
          goto loop;
do_write: Symbol := hd(tl(tl(Instruction)));
          Right := cons(Symbol,tl(Right));
          goto loop;
do_goto:  Nextlabel := hd(tl(tl(Instruction)));
          Qtail := new_tail(Nextlabel, Q);
          goto loop;
do_if:    Symbol := hd(tl(tl(Instruction)));
          Nextlabel := hd(tl(tl(tl(tl(Instruction)))));
          if Symbol = hd(Right) goto jump else loop;
jump: Qtail := new_tail(Nextlabel, Q); 
      goto loop;
error: return list('syntax_error, Instruction);
stop: return Right;`
};

const state = {
    mode: 'generator',
    program: '',
    delta: [],
    args: []
};

function init() {
    const generatorRadio = document.getElementById('mode-generator');
    const evaluatorRadio = document.getElementById('mode-evaluator');
    const paramsInput = document.getElementById('params-input');
    const paramsLabel = document.getElementById('params-label');
    const runBtn = document.getElementById('run-btn');
    const exampleSelect = document.getElementById('example-select');
    const programInput = document.getElementById('program-input');
    const outputDisplay = document.getElementById('output-display');
    const argCountInput = document.getElementById('arg-count');

    generatorRadio.addEventListener('change', () => setMode('generator'));
    evaluatorRadio.addEventListener('change', () => setMode('evaluator'));
    runBtn.addEventListener('click', run);
    exampleSelect.addEventListener('change', loadExample);
    programInput.addEventListener('input', (e) => { state.program = e.target.value; });
    paramsInput.addEventListener('input', handleParamsChange);
    if (argCountInput) {
        argCountInput.addEventListener('input', handleArgCountChange);
    }

    setupOutputSelection();
    updateUI();
}

function setMode(mode) {
    state.mode = mode;
    updateUI();
}

function updateUI() {
    const paramsLabel = document.getElementById('params-label');
    const generatorParams = document.getElementById('generator-params');
    const evaluatorParams = document.getElementById('evaluator-params');

    if (state.mode === 'generator') {
        paramsLabel.textContent = 'Delta (static parameter indices)';
        generatorParams.style.display = 'block';
        evaluatorParams.style.display = 'none';
    } else {
        paramsLabel.textContent = 'Arguments (runtime values)';
        generatorParams.style.display = 'none';
        evaluatorParams.style.display = 'block';
        updateArgInputs();
    }
}

function loadExample() {
    const select = document.getElementById('example-select');
    const exampleKey = select.value;
    const programInput = document.getElementById('program-input');

    if (exampleKey && EXAMPLES[exampleKey]) {
        programInput.value = EXAMPLES[exampleKey];
        state.program = EXAMPLES[exampleKey];
    }

    select.value = '';
}

function handleParamsChange(e) {
    const value = e.target.value.trim();
    state.delta = parseDelta(value);
}

function handleArgCountChange(e) {
    const count = parseInt(e.target.value, 10) || 0;
    state.args = new Array(count).fill('');
    updateArgInputs();
}

function updateArgInputs() {
    const container = document.getElementById('arg-inputs-container');
    const argCountHelp = document.getElementById('arg-count-help');
    const count = parseInt(document.getElementById('arg-count').value, 10) || 0;

    container.innerHTML = '';
    argCountHelp.textContent = `Enter ${count} argument${count !== 1 ? 's' : ''} for the program`;

    for (let i = 0; i < count; i++) {
        const row = document.createElement('div');
        row.className = 'arg-input-row';
        row.innerHTML = `
            <label for="arg-${i}">arg${i}:</label>
            <input type="text" id="arg-${i}" data-index="${i}" placeholder="Enter argument value">
        `;
        container.appendChild(row);

        const input = row.querySelector(`#arg-${i}`);
        input.addEventListener('input', (e) => {
            state.args[parseInt(e.target.dataset.index, 10)] = e.target.value;
        });
    }
}

function parseDelta(value) {
    if (!value.trim()) return [];
    return value.split(',').map(s => {
        const num = parseInt(s.trim(), 10);
        return isNaN(num) ? null : num;
    }).filter(v => v !== null);
}

async function run() {
    const programInput = document.getElementById('program-input');
    const outputDisplay = document.getElementById('output-display');

    const program = programInput.value.trim();
    if (!program) {
        showOutput('Error: Program is required', true);
        return;
    }

    if (state.mode === 'generator') {
        const paramsInput = document.getElementById('params-input');
        handleParamsChange({ target: paramsInput });
    }

    outputDisplay.textContent = 'Processing...';
    outputDisplay.className = '';

    if (state.mode === 'generator') {
        await runGenerator(program);
    } else {
        await runEvaluator(program);
    }
}

async function runGenerator(program) {
    try {
        const response = await fetch('/api/generate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                program: program,
                delta: state.delta
            })
        });

        const data = await response.json();
        if (data.error) {
            showOutput(`Error: ${data.error}`, true);
        } else {
            showOutput(data.result, false);
        }
    } catch (err) {
        showOutput(`Error: ${err.message}`, true);
    }
}

async function runEvaluator(program) {
    try {
        const response = await fetch('/api/evaluate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                program: program,
                args: state.args
            })
        });

        const data = await response.json();
        if (data.error) {
            showOutput(`Error: ${data.error}`, true);
        } else {
            showOutput(`${data.result}`, false);
        }
    } catch (err) {
        showOutput(`Error: ${err.message}`, true);
    }
}

function showOutput(message, isError) {
    const outputDisplay = document.getElementById('output-display');
    outputDisplay.textContent = message;
    outputDisplay.className = isError ? 'error' : 'success';
}

function setupOutputSelection() {
    const outputDisplay = document.getElementById('output-display');
    
    outputDisplay.addEventListener('keydown', (e) => {
        if (e.ctrlKey && e.key === 'a') {
            e.preventDefault();
            const selection = window.getSelection();
            const range = document.createRange();
            range.selectNodeContents(outputDisplay);
            selection.removeAllRanges();
            selection.addRange(range);
        }
    });
}

document.addEventListener('DOMContentLoaded', init);
