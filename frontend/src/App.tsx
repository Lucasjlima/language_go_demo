import { FormEvent, useState } from 'react';
import { ExplainText } from '../wailsjs/go/main/App';

type ExplanationResponse = {
  originalText: string;
  explanation: string;
  tone: string;
  examples: string[];
};

const initialText = 'I cannot believe this is actually working!';

function App() {
  const [text, setText] = useState(initialText);
  const [result, setResult] = useState<ExplanationResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmedText = text.trim();

    if (!trimmedText) {
      setResult(null);
      setErrorMessage('Please enter a phrase before sending it to the backend.');
      return;
    }

    setIsLoading(true);
    setErrorMessage('');

    try {
      const response = await ExplainText(trimmedText);
      setResult(response);
    } catch (error) {
      setResult(null);
      setErrorMessage(error instanceof Error ? error.message : 'Unexpected error while generating explanation.');
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-slate-950 px-5 py-8 text-slate-50">
      <section className="mx-auto w-full max-w-4xl rounded-3xl border border-slate-800 bg-slate-900/80 p-6 shadow-2xl shadow-slate-950/40 backdrop-blur sm:p-8">
        <div className="inline-flex rounded-full bg-sky-500/10 px-3 py-1 text-xs uppercase tracking-[0.18em] text-sky-300">
          Phase 1 frontend validation
        </div>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
          Manual text to explanation
        </h1>
        <p className="mt-3 max-w-3xl text-sm leading-6 text-slate-300 sm:text-base">
          This screen validates the first Wails binding loop: React sends text, Go processes it in
          <code className="rounded bg-slate-800 px-1.5 py-0.5 text-slate-100">internal/ai</code>,
          and the response comes back as a typed object.
        </p>

        <form className="mt-8 grid gap-3" onSubmit={handleSubmit}>
          <label className="text-sm font-medium text-slate-200" htmlFor="text-input">
            Text to explain
          </label>
          <textarea
            className="min-h-36 w-full rounded-2xl border border-slate-700 bg-slate-950/80 px-4 py-3 text-sm leading-6 text-slate-50 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-400/30"
            id="text-input"
            value={text}
            onChange={(event) => setText(event.target.value)}
            placeholder="Paste a sentence, slang, or dialogue snippet"
            rows={5}
          />
          <button
            className="inline-flex w-fit items-center justify-center rounded-full bg-sky-400 px-5 py-3 text-sm font-semibold text-sky-950 transition hover:bg-sky-300 disabled:cursor-wait disabled:bg-sky-200"
            type="submit"
            disabled={isLoading}
          >
            {isLoading ? 'Explaining...' : 'Explain text'}
          </button>
        </form>

        {errorMessage ? (
          <div className="mt-4 rounded-2xl border border-red-400/30 bg-red-950/40 px-4 py-3 text-sm text-red-200">
            {errorMessage}
          </div>
        ) : null}

        <section className="mt-8 border-t border-slate-800 pt-6">
          <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h2 className="text-sm font-medium text-slate-200">Structured response</h2>
            {result ? (
              <span className="inline-flex w-fit rounded-full bg-amber-400/10 px-3 py-1 text-xs font-medium capitalize text-amber-200">
                {result.tone}
              </span>
            ) : null}
          </div>

          {result ? (
            <div className="grid gap-4">
              <div>
                <h3 className="text-sm font-medium text-slate-200">Original text</h3>
                <p className="mt-2 text-sm leading-6 text-slate-300">{result.originalText}</p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-slate-200">Explanation</h3>
                <p className="mt-2 text-sm leading-6 text-slate-300">{result.explanation}</p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-slate-200">Examples</h3>
                <ul className="mt-2 list-disc space-y-2 pl-5 text-sm leading-6 text-slate-300">
                  {result.examples.map((example) => (
                    <li key={example}>{example}</li>
                  ))}
                </ul>
              </div>
            </div>
          ) : (
            <p className="text-sm leading-6 text-slate-400">
              Submit a sentence to see the backend response here.
            </p>
          )}
        </section>
      </section>
    </main>
  );
}

export default App;
