package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"

	"github.com/wirnat/gorviz/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compiles the YAML schema into a static HTML documentation with interactive ERD",
	Long: `The 'compile' command reads the generated YAML file and transforms it into a 
comprehensive static HTML documentation file. This includes both a detailed 
browsable list of models and an interactive Entity-Relationship Diagram (ERD).`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Compiling YAML schema to HTML documentation and ERD...")

		// Read schema.yaml
		yamlFile, err := os.ReadFile("schema.yaml")
		if err != nil {
			fmt.Printf("Error reading schema.yaml. Make sure you've run 'init' command first: %v\n", err)
			os.Exit(1)
		}

		var schema internal.Schema
		err = yaml.Unmarshal(yamlFile, &schema)
		if err != nil {
			fmt.Printf("Error unmarshalling schema.yaml: %v\n", err)
			os.Exit(1)
		}

		// Generate HTML content
		htmlContent, err := generateStaticHTML(&schema)
		if err != nil {
			fmt.Printf("Error generating HTML: %v\n", err)
			os.Exit(1)
		}

		outputFile := "gorviz.html"
		err = os.WriteFile(outputFile, []byte(htmlContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", outputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully generated %s. You can now open it in your browser.\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(compileCmd)
}

type templateData struct {
	Models     []internal.Model
	SchemaJSON template.JS
}

func generateStaticHTML(schema *internal.Schema) (string, error) {
	// Sort models alphabetically by name
	sort.Slice(schema.Models, func(i, j int) bool {
		return schema.Models[i].Name < schema.Models[j].Name
	})

	// Serialize schema to JSON for JS usage
	jsonData, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("doc").Funcs(template.FuncMap{
		"formatTags": func(tags map[string]string) string {
			var parts []string
			for k, v := range tags {
				parts = append(parts, fmt.Sprintf("%s: %s", k, v))
			}
			return strings.Join(parts, " | ")
		},
		"formatType": func(t string) string {
			return strings.TrimPrefix(t, "*")
		},
	}).Parse(docTemplate)

	if err != nil {
		return "", err
	}

	data := templateData{
		Models:     schema.Models,
		SchemaJSON: template.JS(jsonData),
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}

	return b.String(), nil
}

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
            left: 20px;
            z-index: 20;
            display: flex;
            gap: 10px;
        }

        .control-btn {
            background: white;
            border: 1px solid #cbd5e1;
            padding: 8px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.85rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.05);
        }

        .control-btn:hover {
            background: #f8fafc;
        }

    </style>
</head>
<body>

<div class="sidebar">
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

    <div class="physics-controls">
        <button class="control-btn" onclick="network.fit()">Reset View</button>
        <button class="control-btn" onclick="togglePhysics()">Toggle Physics</button>
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
                // title: 'Table: ' + model.table_name, // Replaced by click panel
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
                    let color = '#cbd5e1'; // lighter default
                    
                    visEdges.add({
                        from: model.name,
                        to: target,
                        arrows: { to: { enabled: true, scaleFactor: 0.5 } },
                        color: { color: color, highlight: '#2563eb' },
                        width: 1,
                        selectionWidth: 2,
                        smooth: { type: 'continuous' } // smoother curves
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

        }, 50);
    }

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
