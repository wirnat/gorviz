package cmd

const docTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GORM Schema Visualization</title>
    <script type="text/javascript" src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"></script>
    <style>
        :root {
            --primary-color: #2563eb;
            --sidebar-bg: #f8fafc;
            --sidebar-border: #e2e8f0;
            --text-main: #1e293b;
            --text-secondary: #64748b;
            --bg-main: #ffffff;
            --card-border: #e2e8f0;
            --badge-pk-bg: #fef3c7;
            --badge-pk-text: #92400e;
            --badge-fk-bg: #dbeafe;
            --badge-fk-text: #1e40af;
        }

        * { box-sizing: border-box; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            margin: 0;
            display: flex;
            height: 100vh;
            color: var(--text-main);
            background-color: var(--bg-main);
            overflow: hidden;
        }

        .main-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            margin-left: 0;
        }

        .main-content.expanded {
            margin-left: 0;
        }

        /* Sidebar */
        .sidebar {
            width: 300px;
            background-color: var(--sidebar-bg);
            border-right: 1px solid var(--sidebar-border);
            display: flex;
            flex-direction: column;
            height: 100%;
            z-index: 10;
            flex-shrink: 0;
            transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            position: relative;
        }

        .sidebar.hidden {
            transform: translateX(-100%);
        }

        .sidebar-toggle {
            position: absolute;
            right: -40px;
            top: 20px;
            width: 40px;
            height: 40px;
            background: var(--sidebar-bg);
            border: 1px solid var(--sidebar-border);
            border-left: none;
            border-radius: 0 8px 8px 0;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 11;
            transition: all 0.2s;
        }

        .sidebar-toggle:hover {
            background: #f1f5f9;
            border-color: var(--primary-color);
        }

        .sidebar-toggle::after {
            content: '◀';
            font-size: 14px;
            color: #64748b;
            transition: transform 0.3s;
        }

        .sidebar.hidden .sidebar-toggle::after {
            transform: rotate(180deg);
        }

        .sidebar-header {
            padding: 20px;
            border-bottom: 1px solid var(--sidebar-border);
        }

        .sidebar-header h2 {
            margin: 0 0 15px 0;
            font-size: 1.25rem;
        }

        .view-toggle {
            display: flex;
            gap: 10px;
            margin-bottom: 15px;
        }

        .view-btn {
            flex: 1;
            padding: 8px;
            border: 1px solid var(--primary-color);
            background: white;
            color: var(--primary-color);
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9rem;
            transition: all 0.2s;
        }

        .view-btn.active {
            background: var(--primary-color);
            color: white;
        }

        #searchInput {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid #cbd5e1;
            border-radius: 6px;
            font-size: 0.875rem;
        }

        .nav-list {
            flex: 1;
            overflow-y: auto;
            padding: 10px 0;
            list-style: none;
            margin: 0;
        }

        .nav-item {
            cursor: pointer;
            padding: 8px 20px;
            font-size: 0.9rem;
            transition: background-color 0.2s;
        }

        .nav-item:hover {
            background-color: #e2e8f0;
            color: var(--primary-color);
        }
        
        .nav-item.selected {
            background-color: #e0e7ff;
            color: var(--primary-color);
            font-weight: 600;
        }

        /* Main Content Areas */
        .content-area {
            flex: 1;
            position: relative;
            display: none; /* Hidden by default */
            height: 100%;
            overflow: hidden;
        }

        .content-area.active {
            display: block;
        }

        /* Documentation View */
        #doc-view {
            padding: 40px;
            overflow-y: auto;
            scroll-behavior: smooth;
        }

        .model-card {
            border: 1px solid var(--card-border);
            border-radius: 8px;
            padding: 24px;
            margin-bottom: 40px;
            background: white;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
        }

        .model-header {
            margin-bottom: 20px;
            border-bottom: 2px solid var(--primary-color);
            padding-bottom: 10px;
            display: flex;
            justify-content: space-between;
            align-items: baseline;
        }

        .model-title {
            margin: 0;
            font-size: 1.5rem;
            font-weight: 700;
        }

        .table-name {
            font-family: monospace;
            color: var(--text-secondary);
            font-size: 0.9rem;
        }

        /* Data Tables */
        .data-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.875rem;
            margin-bottom: 24px;
        }

        .data-table th, .data-table td {
            padding: 10px 12px;
            text-align: left;
            border-bottom: 1px solid #f1f5f9;
        }

        .data-table th {
            background-color: #f8fafc;
            font-weight: 600;
            color: var(--text-secondary);
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.05em;
        }

        .data-table tr:hover td {
            background-color: #f8fafc;
        }

        /* Badges & Formatting */
        .badge {
            display: inline-flex;
            align-items: center;
            padding: 2px 8px;
            border-radius: 9999px;
            font-size: 0.7rem;
            font-weight: 600;
            margin-right: 4px;
        }

        .badge-pk {
            background-color: var(--badge-pk-bg);
            color: var(--badge-pk-text);
        }

        .badge-fk {
            background-color: var(--badge-fk-bg);
            color: var(--badge-fk-text);
        }

        .type-text {
            font-family: monospace;
            color: #c026d3;
        }

        .tag-text {
            font-family: monospace;
            color: #64748b;
            font-size: 0.75rem;
            word-break: break-all;
        }

        /* Relationship Legend */
        .relationship-legend {
            position: absolute;
            bottom: 20px;
            left: 20px;
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 15px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            backdrop-filter: blur(5px);
            z-index: 20;
        }

        .relationship-legend h4 {
            margin: 0 0 10px 0;
            font-size: 0.9rem;
            color: #1e293b;
            font-weight: 600;
        }

        .legend-item {
            display: flex;
            align-items: center;
            margin: 6px 0;
            font-size: 0.8rem;
        }

        .legend-line {
            width: 30px;
            height: 2px;
            margin-right: 10px;
            position: relative;
        }

        .legend-line.has-one { background-color: #10b981; }
        .legend-line.has-many { background-color: #3b82f6; }
        .legend-line.belongs-to { background-color: #f59e0b; }
        .legend-line.many-to-many { 
            background-color: #8b5cf6; 
            background-image: repeating-linear-gradient(90deg, #8b5cf6, #8b5cf6 5px, transparent 5px, transparent 10px);
        }
        .legend-line.embedded { 
            background-color: #6b7280; 
            background-image: repeating-linear-gradient(90deg, #6b7280, #6b7280 3px, transparent 3px, transparent 6px);
        }

        .section-header {
            font-size: 1rem;
            font-weight: 600;
            margin: 20px 0 12px 0;
            color: var(--text-main);
        }

        /* Graph View */
        #graph-view {
            background-color: #f0f2f5;
            position: relative;
        }
        
        #network-container {
            width: 100%;
            height: 100%;
            border: none;
        }

        /* Graph Details Sidebar */
        .details-sidebar {
            position: absolute;
            top: 20px;
            right: 20px;
            width: 320px;
            max-height: calc(100% - 40px);
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.1);
            padding: 0;
            display: flex;
            flex-direction: column;
            transform: translateX(350px);
            transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            z-index: 30;
            backdrop-filter: blur(5px);
        }

        .details-sidebar.visible {
            transform: translateX(0);
        }

        .details-header {
            padding: 15px 20px;
            border-bottom: 1px solid #e2e8f0;
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: #fff;
            border-radius: 8px 8px 0 0;
        }

        .details-header h3 {
            margin: 0;
            font-size: 1.1rem;
            color: #1e293b;
        }

        .close-btn {
            background: none;
            border: none;
            font-size: 1.5rem;
            line-height: 1;
            color: #64748b;
            cursor: pointer;
            padding: 0 5px;
        }

        .details-body {
            padding: 0;
            overflow-y: auto;
        }

        .field-list {
            list-style: none;
            padding: 0;
            margin: 0;
        }

        .field-item {
            padding: 10px 20px;
            border-bottom: 1px solid #f1f5f9;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .field-item:last-child {
            border-bottom: none;
        }

        .field-info {
            display: flex;
            flex-direction: column;
        }
        
        .field-name {
            font-weight: 600;
            font-size: 0.9rem;
            color: #334155;
        }

        .field-type {
            font-size: 0.75rem;
            color: #94a3b8;
            font-family: monospace;
            margin-top: 2px;
        }

        .loading-overlay {
            position: absolute;
            top: 0; left: 0; right: 0; bottom: 0;
            background: rgba(255,255,255,0.9);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 20;
            font-size: 1.2rem;
            color: var(--primary-color);
            pointer-events: none;
            opacity: 0;
            transition: opacity 0.3s;
        }
        .loading-overlay.visible {
            opacity: 1;
            pointer-events: all;
        }

        .physics-controls {
            position: absolute;
            bottom: 20px;
            right: 20px;
            z-index: 20;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
            max-width: 300px;
        }

        .control-btn {
            background: white;
            border: 1px solid #cbd5e1;
            padding: 6px 10px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.8rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
            transition: all 0.2s;
        }

        .control-btn:hover {
            background: #f8fafc;
            border-color: #94a3b8;
        }

        .control-btn.active {
            background: var(--primary-color);
            color: white;
            border-color: var(--primary-color);
        }

        /* Enhanced Controls */
        .filter-controls {
            position: absolute;
            top: 20px;
            right: 20px;
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 15px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            backdrop-filter: blur(5px);
            z-index: 25;
            min-width: 200px;
        }

        .filter-controls h4 {
            margin: 0 0 10px 0;
            font-size: 0.9rem;
            color: #1e293b;
            font-weight: 600;
        }

        .filter-item {
            display: flex;
            align-items: center;
            margin: 8px 0;
            font-size: 0.8rem;
        }

        .filter-item input[type="checkbox"] {
            margin-right: 8px;
            cursor: pointer;
        }

        .filter-item label {
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .filter-color {
            width: 12px;
            height: 12px;
            border-radius: 2px;
            border: 1px solid #e2e8f0;
        }

        /* Cardinality Labels */
        .cardinality-label {
            position: absolute;
            background: white;
            border: 1px solid #e2e8f0;
            border-radius: 4px;
            padding: 2px 6px;
            font-size: 10px;
            font-weight: bold;
            font-family: monospace;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            z-index: 15;
            pointer-events: none;
        }

        /* Enhanced Legend */
        .relationship-legend {
            position: absolute;
            bottom: 20px;
            left: 20px;
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 15px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            backdrop-filter: blur(5px);
            z-index: 20;
            max-width: 250px;
        }

        .relationship-legend h4 {
            margin: 0 0 10px 0;
            font-size: 0.9rem;
            color: #1e293b;
            font-weight: 600;
        }

        .legend-item {
            display: flex;
            align-items: center;
            margin: 6px 0;
            font-size: 0.8rem;
            justify-content: space-between;
        }

        .legend-left {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .legend-line {
            width: 25px;
            height: 2px;
            position: relative;
        }

        .legend-line.has-one { background-color: #10b981; }
        .legend-line.has-many { background-color: #3b82f6; }
        .legend-line.belongs-to { background-color: #f59e0b; }
        .legend-line.many-to-many { 
            background-color: #8b5cf6; 
            background-image: repeating-linear-gradient(90deg, #8b5cf6, #8b5cf6 5px, transparent 5px, transparent 10px);
        }
        .legend-line.embedded { 
            background-color: #6b7280; 
            background-image: repeating-linear-gradient(90deg, #6b7280, #6b7280 3px, transparent 3px, transparent 6px);
        }

        .legend-cardinality {
            font-family: monospace;
            font-weight: bold;
            font-size: 11px;
            color: #64748b;
        }

        /* Stats Panel */
        .stats-panel {
            position: absolute;
            top: 20px;
            left: 20px;
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 15px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            backdrop-filter: blur(5px);
            z-index: 25;
            min-width: 200px;
            max-height: 80vh;
            overflow-y: auto;
        }

        .stats-panel h4 {
            margin: 0 0 10px 0;
            font-size: 0.9rem;
            color: #1e293b;
            font-weight: 600;
        }

        .stat-item {
            display: flex;
            justify-content: space-between;
            margin: 6px 0;
            font-size: 0.8rem;
        }

        .stat-value {
            font-weight: 600;
            color: var(--primary-color);
        }

        /* Health Indicators */
        .health-section {
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid #e2e8f0;
        }

        .health-item {
            display: flex;
            align-items: center;
            margin: 8px 0;
            font-size: 0.75rem;
            padding: 6px 8px;
            border-radius: 4px;
            cursor: pointer;
            transition: all 0.2s;
            border: 2px solid transparent;
        }

        .health-item:hover {
            background-color: #f8fafc;
            border-color: #e2e8f0;
            transform: translateX(2px);
        }

        .health-item.active {
            background-color: #eff6ff;
            border-color: var(--primary-color);
            box-shadow: 0 2px 4px rgba(37, 99, 235, 0.1);
        }

        .health-icon {
            width: 16px;
            height: 16px;
            border-radius: 50%;
            margin-right: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 10px;
            font-weight: bold;
            color: white;
        }

        .health-critical { background-color: #dc2626; }
        .health-warning { background-color: #f59e0b; }
        .health-info { background-color: #3b82f6; }
        .health-success { background-color: #10b981; }

        .health-text {
            flex: 1;
            color: #374151;
        }

        .health-count {
            background-color: #f3f4f6;
            padding: 2px 6px;
            border-radius: 10px;
            font-size: 11px;
            font-weight: 600;
            color: #6b7280;
        }

        /* Path Analysis */
        .path-analysis {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: rgba(255, 255, 255, 0.98);
            border: 1px solid #e2e8f0;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
            backdrop-filter: blur(10px);
            z-index: 100;
            min-width: 400px;
            max-width: 600px;
            display: none;
        }

        .path-analysis.visible {
            display: block;
        }

        .path-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }

        .path-title {
            font-size: 1.1rem;
            font-weight: 600;
            color: #1e293b;
        }

        .path-close {
            background: none;
            border: none;
            font-size: 1.5rem;
            color: #6b7280;
            cursor: pointer;
            padding: 0;
        }

        .path-inputs {
            display: flex;
            gap: 10px;
            margin-bottom: 15px;
        }

        .path-input {
            flex: 1;
            padding: 8px 12px;
            border: 1px solid #cbd5e1;
            border-radius: 6px;
            font-size: 0.875rem;
        }

        .path-find-btn {
            background: var(--primary-color);
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.875rem;
            font-weight: 500;
        }

        .path-find-btn:hover {
            background: #1d4ed8;
        }

        .path-results {
            max-height: 200px;
            overflow-y: auto;
        }

        .path-item {
            padding: 10px;
            margin: 8px 0;
            background: #f8fafc;
            border-radius: 6px;
            border-left: 3px solid var(--primary-color);
            font-size: 0.8rem;
        }

        .path-steps {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-top: 5px;
            flex-wrap: wrap;
        }

        .path-step {
            background: white;
            padding: 4px 8px;
            border-radius: 4px;
            border: 1px solid #e2e8f0;
            font-family: monospace;
            font-size: 11px;
        }

        .path-arrow {
            color: #6b7280;
            font-weight: bold;
        }

        /* Loading Optimization */
        .loading-progress {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
            backdrop-filter: blur(5px);
            z-index: 30;
            text-align: center;
            min-width: 250px;
        }

        .progress-bar {
            width: 100%;
            height: 8px;
            background: #f3f4f6;
            border-radius: 4px;
            overflow: hidden;
            margin: 10px 0;
        }

        .progress-fill {
            height: 100%;
            background: var(--primary-color);
            border-radius: 4px;
            transition: width 0.3s ease;
            width: 0%;
        }

        .progress-text {
            font-size: 0.875rem;
            color: #6b7280;
            margin-top: 8px;
        }

        /* Layout Controls */
        .layout-controls {
            position: absolute;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 10px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            backdrop-filter: blur(5px);
            z-index: 20;
            display: flex;
            gap: 8px;
        }

        .layout-btn {
            background: white;
            border: 1px solid #cbd5e1;
            padding: 6px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.75rem;
            transition: all 0.2s;
        }

        .layout-btn:hover {
            background: #f8fafc;
            border-color: #94a3b8;
        }

        .layout-btn.active {
            background: var(--primary-color);
            color: white;
            border-color: var(--primary-color);
        }

        /* Toggle Controls */
        .toggle-controls {
            position: absolute;
            top: 10px;
            right: 10px;
            z-index: 30;
            display: flex;
            gap: 5px;
        }

        .toggle-btn {
            background: rgba(255, 255, 255, 0.9);
            border: 1px solid #e2e8f0;
            border-radius: 4px;
            padding: 4px 8px;
            font-size: 0.7rem;
            cursor: pointer;
            transition: all 0.2s;
            backdrop-filter: blur(5px);
        }

        .toggle-btn:hover {
            background: rgba(255, 255, 255, 1);
            border-color: #cbd5e1;
        }

        .toggle-btn.active {
            background: var(--primary-color);
            color: white;
            border-color: var(--primary-color);
        }

        .panel-hidden {
            display: none !important;
        }

        .panel-minimized {
            transform: scale(0.1);
            opacity: 0;
            pointer-events: none;
        }

        /* Minimized panel indicators */
        .minimized-indicator {
            position: absolute;
            background: var(--primary-color);
            color: white;
            border-radius: 4px;
            padding: 2px 6px;
            font-size: 0.6rem;
            font-weight: bold;
            z-index: 31;
            cursor: pointer;
            transition: all 0.2s;
        }

        .minimized-indicator:hover {
            transform: scale(1.1);
        }

    </style>
</head>
<body>

<div class="sidebar" id="sidebar">
    <div class="sidebar-toggle" onclick="toggleSidebar()" title="Toggle Sidebar"></div>
    <div class="sidebar-header">
        <h2>Schema Viz</h2>
        <div class="view-toggle">
            <button class="view-btn active" onclick="switchView('doc')">Docs</button>
            <button class="view-btn" onclick="switchView('graph')">ERD Graph</button>
        </div>
        <input type="text" id="searchInput" placeholder="Filter models...">
    </div>
    <ul class="nav-list" id="navList">
        {{range .Models}}
        <li class="nav-item" onclick="focusModel('{{.Name}}')">
            {{.Name}}
        </li>
        {{end}}
    </ul>
</div>

<!-- Main Content -->
<div class="main-content" id="main-content">
    <!-- Documentation View -->
    <div id="doc-view" class="content-area active">
    {{range .Models}}
    <div id="model-{{.Name}}" class="model-card">
        <div class="model-header">
            <h3 class="model-title">{{.Name}}</h3>
            <span class="table-name">Table: {{.TableName}}</span>
        </div>

        <div class="section-header">Fields</div>
        <table class="data-table">
            <thead>
                <tr>
                    <th style="width: 20%">Name</th>
                    <th style="width: 15%">Type</th>
                    <th style="width: 10%">Key</th>
                    <th style="width: 55%">Tags / Metadata</th>
                </tr>
            </thead>
            <tbody>
                {{range .Fields}}
                <tr>
                    <td><strong>{{.Name}}</strong></td>
                    <td><span class="type-text">{{formatType .Type}}</span></td>
                    <td>
                        {{if .IsPrimaryKey}}<span class="badge badge-pk">PK</span>{{end}}
                        {{if .IsForeignKey}}<span class="badge badge-fk">FK</span>{{end}}
                    </td>
                    <td><span class="tag-text">{{formatTags .Tags}}</span></td>
                </tr>
                {{end}}
            </tbody>
        </table>

        {{if .Relationships}}
        <div class="section-header">Relationships</div>
        <table class="data-table">
            <thead>
                <tr>
                    <th>Type</th>
                    <th>Target Model</th>
                    <th>Foreign Key</th>
                    <th>References</th>
                </tr>
            </thead>
            <tbody>
                {{range .Relationships}}
                <tr>
                    <td>{{.Type}}</td>
                    <td><span style="color:var(--primary-color); cursor:pointer;" onclick="focusModel('{{.TargetModelName}}')"><strong>{{.TargetModelName}}</strong></span></td>
                    <td class="tag-text">{{.ForeignKey}}</td>
                    <td class="tag-text">{{.References}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{end}}
    </div>
    {{end}}
</div>

<!-- Graph View -->
<div id="graph-view" class="content-area">
    <div id="loading" class="loading-overlay">Generating Layout...</div>
    
    <div id="network-container"></div>

    <!-- Floating Details Panel -->
    <div id="graph-details" class="details-sidebar">
        <div class="details-header">
            <h3 id="detail-title">Model Name</h3>
            <button class="close-btn" onclick="closeDetails()">×</button>
        </div>
        <div class="details-body">
            <ul class="field-list" id="detail-fields">
                <!-- Fields will be injected here -->
            </ul>
        </div>
    </div>

    <!-- Global Toggle Controls -->
    <div class="toggle-controls">
        <button class="toggle-btn" onclick="toggleAllPanels()" title="Toggle All Panels">☰</button>
        <button class="toggle-btn" onclick="togglePanel('stats-panel')" title="Toggle Stats Panel">📊</button>
        <button class="toggle-btn" onclick="togglePanel('filter-controls')" title="Toggle Filters">🔍</button>
        <button class="toggle-btn" onclick="togglePanel('relationship-legend')" title="Toggle Legend">📋</button>
        <button class="toggle-btn" onclick="togglePanel('layout-controls')" title="Toggle Layout">⚙</button>
        <button class="toggle-btn" onclick="togglePanel('physics-controls')" title="Toggle Controls">🎮</button>
    </div>

    <!-- Stats Panel with Health Indicators -->
    <div class="stats-panel" id="stats-panel">
        <h4>Schema Statistics</h4>
        <div class="stat-item">
            <span>Total Models:</span>
            <span class="stat-value" id="total-models">0</span>
        </div>
        <div class="stat-item">
            <span>Total Relationships:</span>
            <span class="stat-value" id="total-relationships">0</span>
        </div>
        <div class="stat-item">
            <span>Visible Models:</span>
            <span class="stat-value" id="visible-models">0</span>
        </div>
        <div class="stat-item">
            <span>Visible Relationships:</span>
            <span class="stat-value" id="visible-relationships">0</span>
        </div>
        
        <div class="health-section">
            <h4>Schema Health</h4>
            <div id="health-indicators">
                <!-- Health indicators will be injected here -->
            </div>
        </div>
    </div>

    <!-- Filter Controls -->
    <div class="filter-controls" id="filter-controls">
        <h4>Filter Relationships</h4>
        <div class="filter-item">
            <input type="checkbox" id="filter-has-one" checked>
            <label for="filter-has-one">
                <div class="filter-color" style="background-color: #10b981;"></div>
                Has One (1:1)
            </label>
        </div>
        <div class="filter-item">
            <input type="checkbox" id="filter-has-many" checked>
            <label for="filter-has-many">
                <div class="filter-color" style="background-color: #3b82f6;"></div>
                Has Many (1:N)
            </label>
        </div>
        <div class="filter-item">
            <input type="checkbox" id="filter-belongs-to" checked>
            <label for="filter-belongs-to">
                <div class="filter-color" style="background-color: #f59e0b;"></div>
                Belongs To (N:1)
            </label>
        </div>
        <div class="filter-item">
            <input type="checkbox" id="filter-many-to-many" checked>
            <label for="filter-many-to-many">
                <div class="filter-color" style="background-color: #8b5cf6;"></div>
                Many to Many (N:M)
            </label>
        </div>
        <div class="filter-item">
            <input type="checkbox" id="filter-embedded" checked>
            <label for="filter-embedded">
                <div class="filter-color" style="background-color: #6b7280;"></div>
                Embedded
            </label>
        </div>
    </div>

    <!-- Enhanced Legend -->
    <div class="relationship-legend" id="relationship-legend">
        <h4>Relationship Types</h4>
        <div class="legend-item">
            <div class="legend-left">
                <div class="legend-line has-one"></div>
                <span>Has One</span>
            </div>
            <span class="legend-cardinality">1:1</span>
        </div>
        <div class="legend-item">
            <div class="legend-left">
                <div class="legend-line has-many"></div>
                <span>Has Many</span>
            </div>
            <span class="legend-cardinality">1:N</span>
        </div>
        <div class="legend-item">
            <div class="legend-left">
                <div class="legend-line belongs-to"></div>
                <span>Belongs To</span>
            </div>
            <span class="legend-cardinality">N:1</span>
        </div>
        <div class="legend-item">
            <div class="legend-left">
                <div class="legend-line many-to-many"></div>
                <span>Many to Many</span>
            </div>
            <span class="legend-cardinality">N:M</span>
        </div>
        <div class="legend-item">
            <div class="legend-left">
                <div class="legend-line embedded"></div>
                <span>Embedded</span>
            </div>
            <span class="legend-cardinality">1:1</span>
        </div>
    </div>

    <!-- Layout Controls -->
    <div class="layout-controls" id="layout-controls">
        <button class="layout-btn active" onclick="changeLayout('physics')">Physics</button>
        <button class="layout-btn" onclick="changeLayout('hierarchical')">Hierarchical</button>
        <button class="layout-btn" onclick="changeLayout('circular')">Circular</button>
        <button class="layout-btn" onclick="changeLayout('clustered')">Clustered</button>
        <button class="layout-btn" onclick="optimizeLayout()">Optimize</button>
    </div>

    <!-- Path Analysis Modal -->
    <div class="path-analysis" id="path-analysis">
        <div class="path-header">
            <h3 class="path-title">Relationship Path Analysis</h3>
            <button class="path-close" onclick="closePathAnalysis()">×</button>
        </div>
        <div class="path-inputs">
            <input type="text" class="path-input" id="path-from" placeholder="From model...">
            <input type="text" class="path-input" id="path-to" placeholder="To model...">
            <button class="path-find-btn" onclick="findPath()">Find Path</button>
        </div>
        <div class="path-results" id="path-results">
            <!-- Path results will be shown here -->
        </div>
    </div>

    <!-- Enhanced Loading -->
    <div class="loading-progress" id="loading-progress" style="display: none;">
        <div>Loading Schema...</div>
        <div class="progress-bar">
            <div class="progress-fill" id="progress-fill"></div>
        </div>
        <div class="progress-text" id="progress-text">Initializing...</div>
    </div>

    <div class="physics-controls" id="physics-controls">
        <button class="control-btn" onclick="network.fit()">Reset View</button>
        <button class="control-btn" onclick="togglePhysics()">Toggle Physics</button>
        <button class="control-btn" onclick="toggleCardinalityLabels()">Cardinality</button>
        <button class="control-btn" onclick="openPathAnalysis()">Path Analysis</button>
        <button class="control-btn" onclick="exportGraph()">Export PNG</button>
    </div>
</div>

<script>
    // Injected Schema Data
    const schemaData = {{.SchemaJSON}};
    
    let network = null;
    let nodes = [];
    let edges = [];
    let currentView = 'doc';
    let physicsEnabled = true;
    let cardinalityLabelsVisible = false;
    let allEdges = [];
    let cardinalityLabels = [];
    let currentLayout = 'physics';
    let healthIssues = [];
    let pathHighlightEdges = [];
    let pathHighlightNodes = [];
    let healthHighlightNodes = [];
    let healthHighlightEdges = [];
    let allPanelsVisible = true;
    let panelStates = {};
    let sidebarVisible = true;

    // Initialize View
    function switchView(view) {
        currentView = view;
        
        // Toggle buttons
        document.querySelectorAll('.view-btn').forEach(btn => btn.classList.remove('active'));
        document.querySelector('button[onclick="switchView(\'' + view + '\')"]').classList.add('active');
        
        // Toggle content
        document.querySelectorAll('.content-area').forEach(area => area.classList.remove('active'));
        document.getElementById(view + '-view').classList.add('active');

        if (view === 'graph') {
            if (!network) {
                initGraph();
            }
        }
    }

    // Focus logic
    function focusModel(modelName) {
        // Update sidebar highlight
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('selected');
            if (item.textContent.trim() === modelName) {
                item.classList.add('selected');
                item.scrollIntoView({ block: 'center', behavior: 'smooth' });
            }
        });

        if (currentView === 'doc') {
            // Scroll doc view
            const el = document.getElementById('model-' + modelName);
            if (el) {
                el.scrollIntoView({ behavior: 'smooth', block: 'start' });
                // Flash effect
                el.style.transition = 'box-shadow 0.3s';
                el.style.boxShadow = '0 0 0 4px rgba(37, 99, 235, 0.2)';
                setTimeout(() => { el.style.boxShadow = ''; }, 1000);
            }
        } else if (currentView === 'graph' && network) {
            // Focus node in graph
            network.selectNodes([modelName]);
            network.focus(modelName, {
                scale: 1.0,
                animation: {
                    duration: 500,
                    easingFunction: 'easeInOutQuad'
                }
            });
            showDetails(modelName);
        }
    }

    function closeDetails() {
        document.getElementById('graph-details').classList.remove('visible');
        if(network) network.unselectAll();
    }

    function showDetails(modelName) {
        const model = schemaData.models.find(m => m.name === modelName);
        if (!model) return;

        document.getElementById('detail-title').textContent = model.name;
        const list = document.getElementById('detail-fields');
        list.innerHTML = '';

        model.fields.forEach(field => {
            const li = document.createElement('li');
            li.className = 'field-item';
            
            let badges = '';
            if (field.is_primary_key) badges += '<span class="badge badge-pk">PK</span>';
            if (field.is_foreign_key) badges += '<span class="badge badge-fk">FK</span>';

            // Cleanup type
            let type = field.type.replace(/^\*/, '');
            if (type.length > 20) type = type.substring(0, 17) + '...';

            li.innerHTML = '<div class="field-info">' +
                '<span class="field-name">' + field.name + '</span>' +
                '<span class="field-type">' + type + '</span>' +
                '</div>' +
                '<div>' + badges + '</div>';
            list.appendChild(li);
        });

        document.getElementById('graph-details').classList.add('visible');
    }

    function togglePhysics() {
        if (!network) return;
        physicsEnabled = !physicsEnabled;
        network.setOptions({ physics: { enabled: physicsEnabled } });
    }

    // Initialize Vis.js Graph
    function initGraph() {
        const container = document.getElementById('network-container');
        const loader = document.getElementById('loading');
        loader.classList.add('visible');

        // Prepare Data
        const visNodes = new vis.DataSet();
        const visEdges = new vis.DataSet();

        schemaData.models.forEach(model => {
            visNodes.add({
                id: model.name,
                label: model.name,
                shape: 'box',
                color: {
                    background: '#ffffff',
                    border: '#64748b',
                    highlight: { background: '#f1f5f9', border: '#2563eb' }
                },
                font: { size: 16, face: 'Segoe UI', color: '#334155' },
                margin: { top: 10, bottom: 10, left: 15, right: 15 },
                borderWidth: 1,
                shadow: { enabled: true, color: 'rgba(0,0,0,0.1)', size: 5, x: 0, y: 2 }
            });

            if (model.relationships) {
                model.relationships.forEach(rel => {
                    let target = rel.target_model_name.replace(/^\*/, '').split('.').pop();
                    
                    let arrows = 'to';
                    let color = '#cbd5e1';
                    let label = '';
                    let dash = false;
                    let fromCardinality = '';
                    let toCardinality = '';
                    
                    // Set different colors and styles based on relationship type
                    switch(rel.type) {
                        case 'has_one':
                            color = '#10b981'; // green
                            arrows = { to: { enabled: true, scaleFactor: 0.6 } };
                            label = 'has one';
                            fromCardinality = '1';
                            toCardinality = '1';
                            break;
                        case 'has_many':
                            color = '#3b82f6'; // blue
                            arrows = { to: { enabled: true, scaleFactor: 0.7 } };
                            label = 'has many';
                            fromCardinality = '1';
                            toCardinality = 'N';
                            break;
                        case 'belongs_to':
                            color = '#f59e0b'; // amber
                            arrows = { to: { enabled: true, scaleFactor: 0.5 } };
                            label = 'belongs to';
                            fromCardinality = 'N';
                            toCardinality = '1';
                            break;
                        case 'many_to_many':
                            color = '#8b5cf6'; // purple
                            arrows = { to: { enabled: true, scaleFactor: 0.6 }, from: { enabled: true, scaleFactor: 0.6 } };
                            label = 'many to many';
                            fromCardinality = 'N';
                            toCardinality = 'M';
                            dash = true;
                            break;
                        case 'embedded':
                            color = '#6b7280'; // gray
                            arrows = { to: { enabled: true, scaleFactor: 0.4 } };
                            label = 'embedded';
                            fromCardinality = '1';
                            toCardinality = '1';
                            dash = true;
                            break;
                        default:
                            color = '#cbd5e1'; // default gray
                            label = rel.type.replace('_', ' ');
                            fromCardinality = '?';
                            toCardinality = '?';
                    }
                    
                    visEdges.add({
                        from: model.name,
                        to: target,
                        arrows: arrows,
                        color: { color: color, highlight: '#2563eb' },
                        width: 2,
                        selectionWidth: 3,
                        smooth: { type: 'continuous' },
                        dashes: dash,
                        label: label,
                        font: {
                            size: 10,
                            align: 'horizontal',
                            background: 'white',
                            strokeWidth: 1,
                            strokeColor: color,
                            face: 'Segoe UI',
                            bold: false
                        },
                        // Store cardinality for later use
                        fromCardinality: fromCardinality,
                        toCardinality: toCardinality,
                        relationshipType: rel.type
                    });
                });
            }
        });

        const data = { nodes: visNodes, edges: visEdges };
        
        // Optimized Physics for large networks
        const options = {
            physics: {
                enabled: true,
                solver: 'barnesHut',
                barnesHut: {
                    gravitationalConstant: -10000, // Strong repulsion
                    centralGravity: 0.1,
                    springLength: 250, // Long edges
                    springConstant: 0.01,
                    damping: 0.15,
                    avoidOverlap: 1 // Prevent nodes covering each other
                },
                stabilization: {
                    enabled: true,
                    iterations: 200, // Pre-calculate layout
                    updateInterval: 20,
                    onlyDynamicEdges: false,
                    fit: true
                }
            },
            interaction: {
                hover: true,
                navigationButtons: false,
                keyboard: false,
                tooltipDelay: 200,
                zoomView: true
            },
            layout: {
                improvedLayout: true
            }
        };

        // Store all edges for filtering
        allEdges = [];
        visEdges.forEach(edge => {
            allEdges.push(edge);
        });

        // Create Network
        setTimeout(() => {
            network = new vis.Network(container, data, options);
            
            network.on("stabilizationIterationsDone", function () {
                network.setOptions( { physics: false } ); // Freeze after load
                physicsEnabled = false;
                loader.classList.remove('visible');
            });

            // Interaction Events
            network.on("selectNode", function (params) {
                if (params.nodes.length > 0) {
                    showDetails(params.nodes[0]);
                }
            });

            network.on("deselectNode", function (params) {
                closeDetails();
            });
            
            network.on("dragStart", function (params) {
                 if (params.nodes.length > 0 && !physicsEnabled) {
                     // Optional: Re-enable physics on drag if desired
                     // network.setOptions({ physics: { enabled: true } });
                 }
            });

            network.on("dragEnd", function (params) {
                if (cardinalityLabelsVisible) {
                    hideCardinalityLabels();
                    setTimeout(showCardinalityLabels, 100);
                }
            });

        }, 50);

        // Add cardinality labels after network is created
        setTimeout(() => {
            updateStats();
            setupFilters();
            analyzeSchemaHealth();
            initializePanelStates();
        }, 100);
    }

    function analyzeSchemaHealth() {
        healthIssues = [];
        
        // Analyze circular dependencies
        const circularDeps = findCircularDependencies();
        if (circularDeps.length > 0) {
            healthIssues.push({
                type: 'critical',
                icon: '!',
                text: 'Circular dependencies detected',
                count: circularDeps.length,
                details: circularDeps
            });
        }

        // Analyze orphaned models
        const orphanedModels = findOrphanedModels();
        if (orphanedModels.length > 0) {
            healthIssues.push({
                type: 'warning',
                icon: '⚠',
                text: 'Orphaned models',
                count: orphanedModels.length,
                details: orphanedModels
            });
        }

        // Analyze deep relationships
        const deepRelationships = findDeepRelationships();
        if (deepRelationships.length > 0) {
            healthIssues.push({
                type: 'info',
                icon: 'i',
                text: 'Deep relationship chains',
                count: deepRelationships.length,
                details: deepRelationships
            });
        }

        // Analyze potential missing indexes
        const missingIndexes = findPotentialMissingIndexes();
        if (missingIndexes.length > 0) {
            healthIssues.push({
                type: 'warning',
                icon: '⚠',
                text: 'Potential missing indexes',
                count: missingIndexes.length,
                details: missingIndexes
            });
        }

        // Render health indicators
        renderHealthIndicators();
    }

    function findCircularDependencies() {
        const visited = new Set();
        const recursionStack = new Set();
        const circularPaths = [];

        function dfs(node, path) {
            if (recursionStack.has(node)) {
                const cycleStart = path.indexOf(node);
                circularPaths.push(path.slice(cycleStart).concat(node));
                return;
            }

            if (visited.has(node)) return;

            visited.add(node);
            recursionStack.add(node);

            const model = schemaData.models.find(m => m.name === node);
            if (model && model.relationships) {
                model.relationships.forEach(rel => {
                    const target = rel.target_model_name.replace(/^\*/, '').split('.').pop();
                    dfs(target, path.concat(node));
                });
            }

            recursionStack.delete(node);
        }

        schemaData.models.forEach(model => {
            if (!visited.has(model.name)) {
                dfs(model.name, []);
            }
        });

        return circularPaths;
    }

    function findOrphanedModels() {
        const connectedModels = new Set();
        
        schemaData.models.forEach(model => {
            if (model.relationships) {
                model.relationships.forEach(rel => {
                    const target = rel.target_model_name.replace(/^\*/, '').split('.').pop();
                    connectedModels.add(target);
                    connectedModels.add(model.name);
                });
            }
        });

        return schemaData.models
            .filter(model => !connectedModels.has(model.name))
            .map(model => model.name);
    }

    function findDeepRelationships() {
        const deepPaths = [];

        function findPathsFrom(node, target, depth, path) {
            if (depth > 5) {
                deepPaths.push(path.concat(target));
                return;
            }

            const model = schemaData.models.find(m => m.name === target);
            if (model && model.relationships) {
                model.relationships.forEach(rel => {
                    const nextTarget = rel.target_model_name.replace(/^\*/, '').split('.').pop();
                    if (nextTarget !== node) {
                        findPathsFrom(node, nextTarget, depth + 1, path.concat(target));
                    }
                });
            }
        }

        schemaData.models.forEach(model => {
            findPathsFrom(model.name, model.name, 0, []);
        });

        return deepPaths;
    }

    function hasIndexTag(tags) {
        // Handle different tag formats
        if (!tags) return false;
        
        if (typeof tags === 'string') {
            return tags.includes('index');
        }
        
        if (Array.isArray(tags)) {
            return tags.includes('index');
        }
        
        if (typeof tags === 'object') {
            return Object.keys(tags).some(key => 
                key === 'index' || tags[key] === 'index'
            );
        }
        
        return false;
    }

    function findPotentialMissingIndexes() {
        const missingIndexes = [];
        
        schemaData.models.forEach(model => {
            if (model.relationships) {
                model.relationships.forEach(rel => {
                    if (rel.foreign_key && rel.type === 'belongs_to') {
                        // Check if foreign key field exists
                        const hasFK = model.fields.some(field => 
                            field.name === rel.foreign_key && 
                            !hasIndexTag(field.tags)
                        );
                        if (hasFK) {
                            missingIndexes.push({
                                model: model.name,
                                field: rel.foreign_key,
                                references: rel.references
                            });
                        }
                    }
                });
            }
        });

        return missingIndexes;
    }

    function renderHealthIndicators() {
        const container = document.getElementById('health-indicators');
        container.innerHTML = '';

        if (healthIssues.length === 0) {
            container.innerHTML = '<div class="health-item"><div class="health-icon health-success">✓</div><div class="health-text">Schema looks healthy!</div></div>';
            return;
        }

        for (let i = 0; i < healthIssues.length; i++) {
            const issue = healthIssues[i];
            const healthItem = document.createElement('div');
            healthItem.className = 'health-item';
            healthItem.id = 'health-item-' + i;
            healthItem.onclick = createHealthClickHandler(issue, i);
            
            const iconDiv = document.createElement('div');
            iconDiv.className = 'health-icon health-' + issue.type;
            iconDiv.textContent = issue.icon;
            
            const textDiv = document.createElement('div');
            textDiv.className = 'health-text';
            textDiv.textContent = issue.text;
            
            const countDiv = document.createElement('div');
            countDiv.className = 'health-count';
            countDiv.textContent = issue.count;
            
            healthItem.appendChild(iconDiv);
            healthItem.appendChild(textDiv);
            healthItem.appendChild(countDiv);
            
            container.appendChild(healthItem);
        }
        
        function createHealthClickHandler(issue, index) {
            return function() { 
                // Clear previous active states
                document.querySelectorAll('.health-item').forEach(item => {
                    item.classList.remove('active');
                });
                
                // Set active state
                const currentItem = document.getElementById('health-item-' + index);
                if (currentItem) {
                    currentItem.classList.add('active');
                }
                
                showHealthDetails(issue); 
            };
        }
    }

    function showHealthDetails(issue) {
        if (issue.details && issue.details.length > 0) {
            highlightHealthIssue(issue);
            
            var detailsText = issue.text + ':\n\n' + issue.details.slice(0, 5).join('\n');
            if (issue.details.length > 5) {
                detailsText += '\n...';
            }
            alert(detailsText);
        }
    }

    function highlightHealthIssue(issue) {
        if (!network) return;
        
        // Clear previous highlights
        clearHealthHighlights();
        
        const nodesToHighlight = new Set();
        const edgesToHighlight = new Set();
        
        switch(issue.type) {
            case 'critical':
                // Circular dependencies - highlight the cycle
                issue.details.forEach(cycle => {
                    cycle.forEach(nodeName => {
                        nodesToHighlight.add(nodeName);
                    });
                    
                    // Find edges in the cycle
                    for (let i = 0; i < cycle.length - 1; i++) {
                        const from = cycle[i];
                        const to = cycle[i + 1];
                        findEdgeBetweenNodes(from, to, edgesToHighlight);
                    }
                });
                break;
                
            case 'warning':
                if (issue.text.includes('Orphaned models')) {
                    // Highlight orphaned models
                    issue.details.forEach(modelName => {
                        nodesToHighlight.add(modelName);
                    });
                } else if (issue.text.includes('missing indexes')) {
                    // Highlight models with missing indexes
                    issue.details.forEach(item => {
                        nodesToHighlight.add(item.model);
                        // Find the relationship edge
                        allEdges.forEach(edge => {
                            if (edge.from === item.model && 
                                edge.relationshipType === 'belongs_to') {
                                edgesToHighlight.add(edge.id);
                            }
                        });
                    });
                }
                break;
                
            case 'info':
                // Deep relationships - highlight the path
                issue.details.forEach(path => {
                    path.forEach(nodeName => {
                        nodesToHighlight.add(nodeName);
                    });
                    
                    // Find edges in the path
                    for (let i = 0; i < path.length - 1; i++) {
                        const from = path[i];
                        const to = path[i + 1];
                        findEdgeBetweenNodes(from, to, edgesToHighlight);
                    }
                });
                break;
        }
        
        // Apply highlights
        if (nodesToHighlight.size > 0) {
            network.selectNodes(Array.from(nodesToHighlight));
        }
        
        if (edgesToHighlight.size > 0) {
            network.selectEdges(Array.from(edgesToHighlight));
        }
        
        // Focus on highlighted elements
        if (nodesToHighlight.size > 0) {
            network.fit({
                nodes: Array.from(nodesToHighlight),
                animation: {
                    duration: 800,
                    easingFunction: 'easeInOutQuad'
                }
            });
        }
        
        // Store for clearing
        healthHighlightNodes = Array.from(nodesToHighlight);
        healthHighlightEdges = Array.from(edgesToHighlight);
    }

    function findEdgeBetweenNodes(from, to, edgesSet) {
        allEdges.forEach(edge => {
            if ((edge.from === from && edge.to === to) || 
                (edge.from === to && edge.to === from)) {
                edgesSet.add(edge.id);
            }
        });
    }

    function clearHealthHighlights() {
        if (network) {
            network.unselectAll();
        }
        healthHighlightNodes = [];
        healthHighlightEdges = [];
    }

    function changeLayout(layoutType) {
        if (!network) return;

        // Update button states
        document.querySelectorAll('.layout-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        event.target.classList.add('active');

        currentLayout = layoutType;
        userPreferences.layout = layoutType;
        saveUserPreferences();

        switch(layoutType) {
            case 'hierarchical':
                network.setOptions({
                    layout: {
                        hierarchical: {
                            direction: 'UD',
                            sortMethod: 'directed',
                            levelSeparation: 150,
                            nodeSpacing: 100
                        }
                    },
                    physics: false
                });
                break;
                
            case 'circular':
                network.setOptions({
                    layout: {
                        improvedLayout: false
                    },
                    physics: {
                        enabled: true,
                        solver: 'forceAtlas2Based',
                        forceAtlas2Based: {
                            gravitationalConstant: -26,
                            centralGravity: 0.01,
                            springLength: 100,
                            avoidOverlap: 1
                        }
                    }
                });
                break;
                
            case 'clustered':
                applyClusteredLayout();
                break;
                
            case 'optimize':
                optimizeLayout();
                break;
                
            default:
                network.setOptions({
                    layout: { improvedLayout: true },
                    physics: { enabled: false }
                });
        }
    }

    function applyClusteredLayout() {
        // Group models by common prefixes
        const clusters = {};
        
        schemaData.models.forEach(model => {
            const prefix = model.name.split(/(?=[A-Z])/)[0];
            if (!clusters[prefix]) clusters[prefix] = [];
            clusters[prefix].push(model.name);
        });

        // Position clusters
        const clusterPositions = {};
        const angleStep = (2 * Math.PI) / Object.keys(clusters).length;
        let angle = 0;

        Object.keys(clusters).forEach(clusterName => {
            const x = Math.cos(angle) * 300;
            const y = Math.sin(angle) * 300;
            clusterPositions[clusterName] = { x, y };
            angle += angleStep;
        });

        // Update node positions
        const nodePositions = {};
        Object.keys(clusters).forEach(clusterName => {
            const cluster = clusters[clusterName];
            const clusterPos = clusterPositions[clusterName];
            
            cluster.forEach((modelName, index) => {
                const angle = (index / cluster.length) * 2 * Math.PI;
                const radius = 50 + (index * 20);
                nodePositions[modelName] = {
                    x: clusterPos.x + Math.cos(angle) * radius,
                    y: clusterPos.y + Math.sin(angle) * radius
                };
            });
        });

        network.moveNode(nodePositions);
        network.setOptions({ physics: false });
    }

    function optimizeLayout() {
        // Show loading
        const loadingEl = document.getElementById('loading-progress');
        const progressFill = document.getElementById('progress-fill');
        const progressText = document.getElementById('progress-text');
        
        loadingEl.style.display = 'block';
        let progress = 0;

        const progressInterval = setInterval(() => {
            progress += 10;
            progressFill.style.width = progress + '%';
            
            if (progress < 30) {
                progressText.textContent = 'Analyzing relationships...';
            } else if (progress < 60) {
                progressText.textContent = 'Optimizing node positions...';
            } else if (progress < 90) {
                progressText.textContent = 'Minimizing edge crossings...';
            } else {
                progressText.textContent = 'Finalizing layout...';
            }
        }, 100);

        setTimeout(() => {
            clearInterval(progressInterval);
            loadingEl.style.display = 'none';
            
            // Apply optimized physics settings
            network.setOptions({
                physics: {
                    enabled: true,
                    solver: 'barnesHut',
                    barnesHut: {
                        gravitationalConstant: -2000,
                        centralGravity: 0.3,
                        springLength: 200,
                        springConstant: 0.04,
                        damping: 0.09,
                        avoidOverlap: 0.1
                    },
                    stabilization: {
                        enabled: true,
                        iterations: 100,
                        updateInterval: 25,
                        onlyDynamicEdges: false,
                        fit: true
                    }
                }
            });

            setTimeout(() => {
                network.setOptions({ physics: false });
            }, 2000);
        }, 1000);
    }

    function openPathAnalysis() {
        document.getElementById('path-analysis').classList.add('visible');
    }

    function closePathAnalysis() {
        document.getElementById('path-analysis').classList.remove('visible');
        clearPathHighlight();
    }

    function findPath() {
        const fromModel = document.getElementById('path-from').value.trim();
        const toModel = document.getElementById('path-to').value.trim();
        
        if (!fromModel || !toModel) {
            alert('Please enter both from and to models');
            return;
        }

        const paths = findAllPaths(fromModel, toModel);
        displayPaths(paths, fromModel, toModel);
    }

    function findAllPaths(from, to) {
        const allPaths = [];
        const visited = new Set();
        
        function dfs(current, target, path) {
            if (current === target) {
                allPaths.push(path.concat(current));
                return;
            }
            
            if (visited.has(current) || path.length > 6) {
                return;
            }
            
            visited.add(current);
            
            const model = schemaData.models.find(m => m.name === current);
            if (model && model.relationships) {
                model.relationships.forEach(rel => {
                    const next = rel.target_model_name.replace(/^\*/, '').split('.').pop();
                    dfs(next, target, path.concat(current));
                });
            }
            
            visited.delete(current);
        }
        
        dfs(from, to, []);
        return allPaths.slice(0, 5); // Limit to 5 shortest paths
    }

    function displayPaths(paths, from, to) {
        const resultsEl = document.getElementById('path-results');
        
        if (paths.length === 0) {
            resultsEl.innerHTML = '<div class="path-item">No path found between models</div>';
            return;
        }
        
        var html = '';
        for (var i = 0; i < paths.length; i++) {
            var path = paths[i];
            var pathHtml = '<div class="path-item" onclick="highlightPath(' + i + ')"><strong>Path ' + (i + 1) + '</strong> (' + (path.length - 1) + ' steps)<div class="path-steps">';
            
            for (var j = 0; j < path.length; j++) {
                pathHtml += '<span class="path-step">' + path[j] + '</span>';
                if (j < path.length - 1) {
                    pathHtml += '<span class="path-arrow">→</span>';
                }
            }
            
            pathHtml += '</div></div>';
            html += pathHtml;
        }
        resultsEl.innerHTML = html;
        
        // Store paths for highlighting
        window.foundPaths = paths;
    }

    function highlightPath(pathIndex) {
        clearPathHighlight();
        
        if (!window.foundPaths || !window.foundPaths[pathIndex]) return;
        
        const path = window.foundPaths[pathIndex];
        const nodesToHighlight = new Set(path);
        const edgesToHighlight = new Set();
        
        // Find edges in the path
        for (let i = 0; i < path.length - 1; i++) {
            const from = path[i];
            const to = path[i + 1];
            
            allEdges.forEach(edge => {
                if ((edge.from === from && edge.to === to) || 
                    (edge.from === to && edge.to === from)) {
                    edgesToHighlight.add(edge.id);
                }
            });
        }
        
        // Highlight nodes
        network.selectNodes(Array.from(nodesToHighlight));
        
        // Highlight edges
        network.selectEdges(Array.from(edgesToHighlight));
        
        // Store for clearing
        pathHighlightNodes = Array.from(nodesToHighlight);
        pathHighlightEdges = Array.from(edgesToHighlight);
        
        // Focus on the path
        network.fit({
            nodes: Array.from(nodesToHighlight),
            animation: true
        });
    }

    function clearPathHighlight() {
        if (network) {
            network.unselectAll();
        }
        pathHighlightNodes = [];
        pathHighlightEdges = [];
    }

    function toggleAllPanels() {
        allPanelsVisible = !allPanelsVisible;
        const panels = ['stats-panel', 'filter-controls', 'relationship-legend', 'layout-controls', 'physics-controls'];
        
        panels.forEach(panelId => {
            const panel = document.getElementById(panelId);
            if (panel) {
                if (allPanelsVisible) {
                    panel.classList.remove('panel-hidden');
                    panel.classList.remove('panel-minimized');
                } else {
                    panel.classList.add('panel-hidden');
                }
                panelStates[panelId] = allPanelsVisible;
            }
        });

        // Update toggle button states
        updateToggleButtons();
    }

    function togglePanel(panelId) {
        const panel = document.getElementById(panelId);
        if (!panel) return;

        const currentState = panelStates[panelId] !== false; // Default to true
        const newState = !currentState;
        
        if (newState) {
            panel.classList.remove('panel-hidden');
            panel.classList.remove('panel-minimized');
        } else {
            panel.classList.add('panel-hidden');
        }
        
        panelStates[panelId] = newState;
        userPreferences.panelStates[panelId] = newState;
        updateToggleButtons();
        
        // Save preferences
        saveUserPreferences();
        
        // Check if all panels are hidden
        checkAllPanelsState();
    }

    function updateToggleButtons() {
        const panels = ['stats-panel', 'filter-controls', 'relationship-legend', 'layout-controls', 'physics-controls'];
        const toggleBtns = document.querySelectorAll('.toggle-btn');
        
        panels.forEach((panelId, index) => {
            if (toggleBtns[index + 1]) { // Skip first button (toggle all)
                const isVisible = panelStates[panelId] !== false;
                if (isVisible) {
                    toggleBtns[index + 1].classList.add('active');
                } else {
                    toggleBtns[index + 1].classList.remove('active');
                }
            }
        });

        // Update toggle all button
        const toggleAllBtn = document.querySelector('.toggle-btn');
        if (toggleAllBtn) {
            const allVisible = panels.every(panelId => panelStates[panelId] !== false);
            if (allVisible) {
                toggleAllBtn.classList.add('active');
                toggleAllBtn.textContent = '☰';
            } else {
                toggleAllBtn.classList.remove('active');
                toggleAllBtn.textContent = '☰';
            }
        }
    }

    function checkAllPanelsState() {
        const panels = ['stats-panel', 'filter-controls', 'relationship-legend', 'layout-controls', 'physics-controls'];
        const visibleCount = panels.filter(panelId => panelStates[panelId] !== false).length;
        
        allPanelsVisible = visibleCount === panels.length;
    }

    function showMinimizedIndicator(panelId, text) {
        // Remove existing indicator for this panel
        const existingIndicator = document.getElementById('minimized-' + panelId);
        if (existingIndicator) {
            existingIndicator.remove();
        }

        // Create new indicator
        const indicator = document.createElement('div');
        indicator.id = 'minimized-' + panelId;
        indicator.className = 'minimized-indicator';
        indicator.textContent = text;
        indicator.onclick = function() { togglePanel(panelId); };
        
        // Position based on panel type
        const panel = document.getElementById(panelId);
        if (panel) {
            const rect = panel.getBoundingClientRect();
            indicator.style.left = rect.left + 'px';
            indicator.style.top = rect.top + 'px';
        }
        
        document.body.appendChild(indicator);
    }

    function hideMinimizedIndicator(panelId) {
        const indicator = document.getElementById('minimized-' + panelId);
        if (indicator) {
            indicator.remove();
        }
    }

    // Initialize panel states
    function initializePanelStates() {
        // Load saved preferences
        loadUserPreferences();
        
        const panels = ['stats-panel', 'filter-controls', 'relationship-legend', 'layout-controls', 'physics-controls'];
        panels.forEach(panelId => {
            // Use saved state or default to true
            panelStates[panelId] = userPreferences.panelStates[panelId] !== false;
        });
        
        // Apply saved filters
        if (userPreferences.filters) {
            Object.keys(userPreferences.filters).forEach(filterType => {
                const checkbox = document.getElementById('filter-' + filterType.replace('_', '-'));
                if (checkbox) {
                    checkbox.checked = userPreferences.filters[filterType];
                }
            });
        }
        
        // Apply saved layout
        if (userPreferences.layout && userPreferences.layout !== 'physics') {
            setTimeout(() => {
                changeLayout(userPreferences.layout);
            }, 500);
        }
        
        // Apply saved sidebar state
        sidebarVisible = userPreferences.sidebarVisible !== false;
        if (!sidebarVisible) {
            document.getElementById('sidebar').classList.add('hidden');
            document.getElementById('main-content').classList.add('expanded');
        }
        
        updateToggleButtons();
    }

    function toggleSidebar() {
        const sidebar = document.getElementById('sidebar');
        const mainContent = document.getElementById('main-content');
        
        sidebarVisible = !sidebarVisible;
        
        if (sidebarVisible) {
            sidebar.classList.remove('hidden');
            mainContent.classList.remove('expanded');
        } else {
            sidebar.classList.add('hidden');
            mainContent.classList.add('expanded');
        }
        
        // Save preference
        userPreferences.sidebarVisible = sidebarVisible;
        saveUserPreferences();
        
        // Resize network if exists
        if (network && currentView === 'graph') {
            setTimeout(() => {
                network.fit();
            }, 300);
        }
    }

    // User preferences management
    let userPreferences = {
        layout: 'physics',
        filters: {},
        panelStates: {},
        recentSearches: []
    };

    function loadUserPreferences() {
        try {
            const saved = localStorage.getItem('erd-preferences');
            if (saved) {
                userPreferences = JSON.parse(saved);
            }
        } catch (e) {
            console.log('Using default preferences');
        }
    }

    function saveUserPreferences() {
        try {
            localStorage.setItem('erd-preferences', JSON.stringify(userPreferences));
        } catch (e) {
            console.log('Failed to save preferences');
        }
    }

    function updatePreference(key, value) {
        userPreferences[key] = value;
        saveUserPreferences();
    }

    function updateStats() {
        const totalModels = schemaData.models.length;
        let totalRelationships = 0;
        
        schemaData.models.forEach(model => {
            if (model.relationships) {
                totalRelationships += model.relationships.length;
            }
        });

        document.getElementById('total-models').textContent = totalModels;
        document.getElementById('total-relationships').textContent = totalRelationships;
        document.getElementById('visible-models').textContent = totalModels;
        document.getElementById('visible-relationships').textContent = totalRelationships;
    }

    function setupFilters() {
        const filters = ['has-one', 'has-many', 'belongs-to', 'many-to-many', 'embedded'];
        
        filters.forEach(filterType => {
            const checkbox = document.getElementById('filter-' + filterType);
            if (checkbox) {
                checkbox.addEventListener('change', applyFilters);
            }
        });
    }

    function applyFilters() {
        const filters = {
            'has_one': document.getElementById('filter-has-one').checked,
            'has_many': document.getElementById('filter-has-many').checked,
            'belongs_to': document.getElementById('filter-belongs-to').checked,
            'many_to_many': document.getElementById('filter-many-to-many').checked,
            'embedded': document.getElementById('filter-embedded').checked
        };

        // Save filter preferences
        userPreferences.filters = filters;
        saveUserPreferences();

        const edgesToShow = [];
        const edgesToHide = [];

        allEdges.forEach(edge => {
            const relType = edge.relationshipType;
            if (filters[relType]) {
                edgesToShow.push(edge.id);
            } else {
                edgesToHide.push(edge.id);
            }
        });

        if (network) {
            // Hide edges first
            edgesToHide.forEach(edgeId => {
                network.body.data.edges.update({ 
                    id: edgeId, 
                    hidden: true 
                });
            });

            // Show edges
            edgesToShow.forEach(edgeId => {
                network.body.data.edges.update({ 
                    id: edgeId, 
                    hidden: false 
                });
            });

            // Update visible count
            const visibleCount = edgesToShow.length;
            document.getElementById('visible-relationships').textContent = visibleCount;
        }
    }

    function toggleCardinalityLabels() {
        if (!network) return;
        
        cardinalityLabelsVisible = !cardinalityLabelsVisible;
        
        if (cardinalityLabelsVisible) {
            showCardinalityLabels();
        } else {
            hideCardinalityLabels();
        }
    }

    function showCardinalityLabels() {
        const container = document.getElementById('network-container');
        const positions = network.getPositions();
        
        allEdges.forEach(edge => {
            const fromPos = positions[edge.from];
            const toPos = positions[edge.to];
            
            if (fromPos && toPos) {
                const midX = (fromPos.x + toPos.x) / 2;
                const midY = (fromPos.y + toPos.y) / 2;
                
                // Create from cardinality label
                const fromLabel = document.createElement('div');
                fromLabel.className = 'cardinality-label';
                fromLabel.textContent = edge.fromCardinality;
                fromLabel.style.left = (fromPos.x + (midX - fromPos.x) * 0.3) + 'px';
                fromLabel.style.top = (fromPos.y + (midY - fromPos.y) * 0.3) + 'px';
                fromLabel.id = 'cardinality-from-' + edge.id;
                
                // Create to cardinality label
                const toLabel = document.createElement('div');
                toLabel.className = 'cardinality-label';
                toLabel.textContent = edge.toCardinality;
                toLabel.style.left = (toPos.x - (toPos.x - midX) * 0.3) + 'px';
                toLabel.style.top = (toPos.y - (toPos.y - midY) * 0.3) + 'px';
                toLabel.id = 'cardinality-to-' + edge.id;
                
                container.appendChild(fromLabel);
                container.appendChild(toLabel);
                cardinalityLabels.push(fromLabel, toLabel);
            }
        });
    }

    function hideCardinalityLabels() {
        cardinalityLabels.forEach(label => {
            if (label.parentNode) {
                label.parentNode.removeChild(label);
            }
        });
        cardinalityLabels = [];
    }

    function exportGraph() {
        if (!network) return;
        
        // Create canvas
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        
        // Get network canvas
        const networkCanvas = document.querySelector('#network-container canvas');
        if (!networkCanvas) return;
        
        canvas.width = networkCanvas.width;
        canvas.height = networkCanvas.height;
        
        // Draw white background
        ctx.fillStyle = 'white';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // Draw network
        ctx.drawImage(networkCanvas, 0, 0);
        
        // Download
        const link = document.createElement('a');
        link.download = 'gorm-schema-erd.png';
        link.href = canvas.toDataURL();
        link.click();
    }

    // Keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K: Quick toggle all panels
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            toggleAllPanels();
        }
        
        // Ctrl/Cmd + H: Toggle health panel
        if ((e.ctrlKey || e.metaKey) && e.key === 'h') {
            e.preventDefault();
            togglePanel('stats-panel');
        }
        
        // Ctrl/Cmd + F: Toggle filter panel
        if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
            e.preventDefault();
            togglePanel('filter-controls');
        }
        
        // Ctrl/Cmd + L: Toggle legend
        if ((e.ctrlKey || e.metaKey) && e.key === 'l') {
            e.preventDefault();
            togglePanel('relationship-legend');
        }
        
        // Ctrl/Cmd + B: Toggle sidebar
        if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
            e.preventDefault();
            toggleSidebar();
        }

        // Escape: Close all modals and show all panels
        if (e.key === 'Escape') {
            closePathAnalysis();
            clearHealthHighlights();
            if (!allPanelsVisible) {
                toggleAllPanels();
            }
        }
    });

    // Search Logic
    document.getElementById('searchInput').addEventListener('keyup', function() {
        const filter = this.value.toLowerCase();
        const navItems = document.querySelectorAll('.nav-item');
        
        navItems.forEach(item => {
            const text = item.textContent.toLowerCase();
            if (text.includes(filter)) {
                item.style.display = '';
            } else {
                item.style.display = 'none';
            }
        });
    });

</script>

</body>
</html>
`
