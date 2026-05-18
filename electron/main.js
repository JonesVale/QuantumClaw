const {
  app,
  BrowserWindow,
  Menu,
  Tray,
  nativeImage,
  nativeTheme,
  globalShortcut,
  shell,
} = require('electron');
const path = require('path');
const fs = require('fs');
const { autoUpdater } = require('electron-updater');

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const PRELOAD_PATH = path.join(__dirname, 'preload.js');
const LOGO_PATH = path.join(__dirname, '..', 'web', 'default', 'public', 'logo.webp');

const WINDOW_STATE_FILE = 'window-state.json';
const WINDOW_DEFAULTS = {
  width: 1400,
  height: 900,
  minWidth: 1024,
  minHeight: 680,
};

const UPDATE_INTERVAL_MS = 1000 * 60 * 60; // check every hour

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

let mainWindow = null;
let tray = null;
let isQuitting = false;

// ---------------------------------------------------------------------------
// Window state persistence
// ---------------------------------------------------------------------------

function windowStatePath() {
  return path.join(app.getPath('userData'), WINDOW_STATE_FILE);
}

function loadWindowState() {
  try {
    const raw = fs.readFileSync(windowStatePath(), 'utf-8');
    return { ...WINDOW_DEFAULTS, ...JSON.parse(raw) };
  } catch {
    return { ...WINDOW_DEFAULTS };
  }
}

function saveWindowState() {
  if (!mainWindow) return;
  try {
    const bounds = mainWindow.getBounds();
    const maximized = mainWindow.isMaximized();
    fs.writeFileSync(windowStatePath(), JSON.stringify({ ...bounds, maximized }, null, 2));
  } catch {
    // best-effort
  }
}

// ---------------------------------------------------------------------------
// Auto-updater
// ---------------------------------------------------------------------------

function setupAutoUpdater() {
  autoUpdater.autoDownload = false;
  autoUpdater.autoInstallOnAppQuit = true;

  autoUpdater.on('update-available', (info) => {
    const dialog = require('electron').dialog;
    dialog
      .showMessageBox(mainWindow, {
        type: 'info',
        title: 'Update Available',
        message: `A new version (${info.version}) is available.`,
        detail: 'Do you want to download and install it now?',
        buttons: ['Download', 'Later'],
        defaultId: 0,
        cancelId: 1,
      })
      .then(({ response }) => {
        if (response === 0) autoUpdater.downloadUpdate();
      });
  });

  autoUpdater.on('update-downloaded', () => {
    const dialog = require('electron').dialog;
    dialog
      .showMessageBox(mainWindow, {
        type: 'info',
        title: 'Update Ready',
        message: 'Update downloaded. Restart to install?',
        buttons: ['Restart', 'Later'],
        defaultId: 0,
        cancelId: 1,
      })
      .then(({ response }) => {
        if (response === 0) autoUpdater.quitAndInstall();
      });
  });

  autoUpdater.on('error', (err) => {
    console.error('[auto-updater]', err.message);
  });

  // Check for updates periodically
  setInterval(() => {
    autoUpdater.checkForUpdates().catch(() => {});
  }, UPDATE_INTERVAL_MS);

  // Also check once shortly after startup
  app.whenReady().then(() => {
    // Delay the first check so the window is ready first
    setTimeout(() => {
      autoUpdater.checkForUpdates().catch(() => {});
    }, 10_000);
  });
}

// ---------------------------------------------------------------------------
// Dark mode
// ---------------------------------------------------------------------------

function applyTheme() {
  if (!mainWindow) return;
  const shouldBeDark = nativeTheme.shouldUseDarkColors;
  mainWindow.webContents.send('theme-changed', shouldBeDark ? 'dark' : 'light');
}

// ---------------------------------------------------------------------------
// Tray
// ---------------------------------------------------------------------------

function createTray() {
  let trayIcon;
  try {
    trayIcon = nativeImage.createFromPath(LOGO_PATH);
    // Resize to a typical tray size (16x16 on most platforms, 22x22 on Linux)
    trayIcon = trayIcon.resize({ width: 16, height: 16 });
  } catch {
    // Fallback: create an empty 16x16 image
    trayIcon = nativeImage.createEmpty();
  }

  tray = new Tray(trayIcon);
  tray.setToolTip('QuantumClaw');

  const contextMenu = Menu.buildFromTemplate([
    {
      label: 'Show / Hide',
      click: () => {
        if (mainWindow) {
          if (mainWindow.isVisible()) {
            mainWindow.hide();
          } else {
            mainWindow.show();
            mainWindow.focus();
          }
        }
      },
    },
    { type: 'separator' },
    {
      label: 'Quit',
      click: () => {
        isQuitting = true;
        app.quit();
      },
    },
  ]);

  tray.setContextMenu(contextMenu);

  // Click on tray icon toggles window visibility (except on macOS where
  // clicking the dock icon already handles this)
  tray.on('click', () => {
    if (mainWindow) {
      if (mainWindow.isVisible()) {
        mainWindow.hide();
      } else {
        mainWindow.show();
        mainWindow.focus();
      }
    }
  });
}

// ---------------------------------------------------------------------------
// Keyboard shortcuts
// ---------------------------------------------------------------------------

function registerGlobalShortcuts() {
  // CmdOrCtrl+Q → quit (override Electron's default to add our quit flag)
  globalShortcut.register('CmdOrCtrl+Q', () => {
    isQuitting = true;
    app.quit();
  });

  // CmdOrCtrl+R → reload the window content
  globalShortcut.register('CmdOrCtrl+R', () => {
    if (mainWindow) mainWindow.webContents.reload();
  });
}

// ---------------------------------------------------------------------------
// Application menu
// ---------------------------------------------------------------------------

function setupMenu() {
  const template = [
    {
      label: 'QuantumClaw',
      submenu: [
        { role: 'about', label: 'About QuantumClaw' },
        { type: 'separator' },
        { role: 'quit', label: 'Quit' },
      ],
    },
    {
      label: 'Edit',
      submenu: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        { role: 'selectAll' },
      ],
    },
    {
      label: 'View',
      submenu: [
        { role: 'reload' },
        { role: 'forceReload' },
        { role: 'toggleDevTools' },
        { type: 'separator' },
        { role: 'resetZoom' },
        { role: 'zoomIn' },
        { role: 'zoomOut' },
        { type: 'separator' },
        { role: 'togglefullscreen' },
      ],
    },
    {
      label: 'Window',
      submenu: [
        { role: 'minimize' },
        { role: 'zoom' },
        { role: 'close' },
      ],
    },
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}

// ---------------------------------------------------------------------------
// Window creation
// ---------------------------------------------------------------------------

function createWindow() {
  const state = loadWindowState();

  mainWindow = new BrowserWindow({
    ...state,
    minWidth: WINDOW_DEFAULTS.minWidth,
    minHeight: WINDOW_DEFAULTS.minHeight,
    icon: LOGO_PATH,
    webPreferences: {
      preload: PRELOAD_PATH,
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
    },
    titleBarStyle: 'hiddenInset',
    show: false,
  });

  // Restore maximized state
  if (state.maximized) {
    mainWindow.maximize();
  }

  // Dev mode: load from local server, Prod mode: load from bundled build
  const isDev = process.env.NODE_ENV === 'development';
  const serverUrl = process.env.QC_SERVER_URL || 'http://localhost:3666';

  if (isDev) {
    mainWindow.loadURL(serverUrl);
    mainWindow.webContents.openDevTools();
  } else {
    mainWindow.loadURL(serverUrl);
  }

  mainWindow.once('ready-to-show', () => {
    mainWindow.show();
  });

  // Save state on resize / move
  mainWindow.on('resize', saveWindowState);
  mainWindow.on('move', saveWindowState);
  mainWindow.on('maximize', saveWindowState);
  mainWindow.on('unmaximize', saveWindowState);

  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  // Close-to-tray behaviour (on Cmd+W / window close button)
  mainWindow.on('close', (event) => {
    if (!isQuitting) {
      event.preventDefault();
      mainWindow.hide();
    }
  });

  // Open external links in the default browser
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: 'deny' };
  });

  setupMenu();
  applyTheme();
}

// ---------------------------------------------------------------------------
// App lifecycle
// ---------------------------------------------------------------------------

app.whenReady().then(() => {
  createTray();
  registerGlobalShortcuts();
  createWindow();
  setupAutoUpdater();
});

// Listen for system theme changes
nativeTheme.on('updated', applyTheme);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    isQuitting = true;
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  } else {
    // Re-show on macOS dock click
    mainWindow?.show();
  }
});

app.on('before-quit', () => {
  isQuitting = true;
  saveWindowState();
  globalShortcut.unregisterAll();
});

app.on('will-quit', () => {
  globalShortcut.unregisterAll();
});
