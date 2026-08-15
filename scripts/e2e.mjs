import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const repositoryDir = resolve(scriptDir, "..");
const frontendDir = resolve(repositoryDir, "frontend");
const composeFile = resolve(repositoryDir, "docker-compose.e2e.yml");
const resultsDir = resolve(frontendDir, "test-results");
const [action = "run", ...playwrightArgs] = process.argv.slice(2);
const supportedActions = new Set(["run", "up", "down", "logs"]);

if (!supportedActions.has(action)) {
  throw new Error(
    `Usage: node scripts/e2e.mjs {run|up|down|logs} [playwright arguments]`,
  );
}

const projectName =
  process.env.E2E_COMPOSE_PROJECT ?? `go-infra-link-e2e-${process.pid}`;
const composeArgs = [
  "compose",
  "--project-name",
  projectName,
  "--file",
  composeFile,
];
let cleanupPromise;
let interruptedBy = null;

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.once(signal, () => {
    interruptedBy ??= signal;
    process.exitCode = signal === "SIGINT" ? 130 : 143;
    void cleanup().catch((error) => {
      console.error(`E2E cleanup after ${signal} failed:`, error);
    });
  });
}

try {
  switch (action) {
    case "run":
      await run();
      break;
    case "up":
      await compose(["up", "--build", "--wait", "--wait-timeout", "180"]);
      process.stdout.write(`${await frontendAddress()}\n`);
      break;
    case "down":
      await cleanup();
      break;
    case "logs":
      await compose(["logs", "--no-color"]);
      break;
  }
} catch (error) {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode ||= 1;
}

async function run() {
  let primaryError = null;
  try {
    await compose(["up", "--build", "--wait", "--wait-timeout", "180"]);
    throwIfInterrupted();

    const address = await frontendAddress();
    const port = extractPort(address);
    await command("pnpm", ["exec", "playwright", "test", ...playwrightArgs], {
      cwd: frontendDir,
      env: {
        ...process.env,
        E2E_BASE_URL: `http://127.0.0.1:${port}`,
        E2E_COMPOSE_PROJECT: projectName,
      },
    });
    throwIfInterrupted();
  } catch (error) {
    primaryError = error;
  }

  try {
    await cleanup();
  } catch (cleanupError) {
    if (primaryError) {
      console.error(
        "E2E cleanup failed after the primary error:",
        cleanupError,
      );
    } else {
      primaryError = cleanupError;
    }
  }

  if (primaryError) throw primaryError;
}

function cleanup() {
  cleanupPromise ??= cleanupStack();
  return cleanupPromise;
}

async function cleanupStack() {
  await saveComposeLogs();
  await compose(["down", "--volumes", "--remove-orphans"]);
}

async function saveComposeLogs() {
  await mkdir(resultsDir, { recursive: true });
  const result = await captureCompose(["logs", "--no-color"]);
  await writeFile(resolve(resultsDir, "compose.log"), result.output, "utf8");
  if (result.code !== 0) {
    console.warn(`Could not collect compose logs (exit ${result.code}).`);
  }
}

async function frontendAddress() {
  const result = await captureCompose(["port", "frontend", "80"]);
  if (result.code !== 0 || !result.output.trim()) {
    throw new Error(
      "Could not resolve the dynamically published frontend test port.",
    );
  }
  return result.output.trim().split(/\r?\n/u)[0];
}

function extractPort(address) {
  const match = /:(\d+)$/u.exec(address);
  if (!match) {
    throw new Error(`Could not parse the frontend test port from ${address}.`);
  }
  return match[1];
}

function compose(args) {
  return command("docker", [...composeArgs, ...args], { cwd: repositoryDir });
}

function captureCompose(args) {
  return captureCommand("docker", [...composeArgs, ...args], {
    cwd: repositoryDir,
  });
}

async function command(executable, args, options) {
  const result = await captureCommand(executable, args, {
    ...options,
    inherit: true,
  });
  if (result.code !== 0) {
    throw new Error(
      `${executable} ${args.join(" ")} failed with exit code ${result.code}.`,
    );
  }
}

function captureCommand(executable, args, { cwd, env, inherit = false }) {
  return new Promise((resolvePromise, reject) => {
    const child = spawn(executable, args, {
      cwd,
      env,
      shell: process.platform === "win32",
      stdio: inherit ? "inherit" : ["ignore", "pipe", "pipe"],
    });
    let output = "";

    if (!inherit) {
      child.stdout.on("data", (chunk) => {
        output += chunk;
      });
      child.stderr.on("data", (chunk) => {
        output += chunk;
      });
    }

    child.once("error", reject);
    child.once("close", (code) => {
      resolvePromise({ code: code ?? 1, output });
    });
  });
}

function throwIfInterrupted() {
  if (interruptedBy) {
    throw new Error(`E2E run interrupted by ${interruptedBy}.`);
  }
}
