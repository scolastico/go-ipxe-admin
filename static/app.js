function appData() {
    return {
        tab: 'deployments',
        deployments: [],
        templates: [],
        assets: [],
        recentClients: [],
        modals: {
            deployment: false,
            template: false,
            asset: false
        },
        forms: {
            deployment: { ip: '', template_id: '', variables: {}, timeout: 0, once: false },
            template: { name: '', variables: [], content: '' },
            asset: { filename: '', content: '' }
        },
        
        async initApp() {
            await this.fetchData();
        },

        async fetchData() {
            this.deployments = await (await fetch('/api/deployments')).json() || [];
            this.templates = await (await fetch('/api/templates')).json() || [];
            this.assets = await (await fetch('/api/assets')).json() || [];
            this.recentClients = await (await fetch('/api/clients/recent')).json() || [];
        },

        getTemplateName(id) {
            const t = this.templates.find(x => x.id === id);
            return t ? t.name : 'Unknown';
        },

        getTemplateVars(id) {
            const t = this.templates.find(x => x.id === id);
            return t ? t.variables || [] : [];
        },

        openDeploymentModal() {
            this.forms.deployment = { ip: '', template_id: '', variables: {}, timeout: 0, once: false };
            this.modals.deployment = true;
        },

        editDeployment(d) {
            this.forms.deployment = JSON.parse(JSON.stringify(d));
            this.modals.deployment = true;
        },

        updateDeploymentVars() {
            const vars = this.getTemplateVars(this.forms.deployment.template_id);
            const newVars = {};
            vars.forEach(v => {
                if (v.default) {
                    newVars[v.name] = v.default;
                } else if (v.type === 'select' && v.options) {
                    newVars[v.name] = v.options.split(',')[0].trim();
                } else {
                    newVars[v.name] = '';
                }
            });
            this.forms.deployment.variables = newVars;
        },

        async saveDeployment() {
            await fetch('/api/deployments', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.forms.deployment)
            });
            this.modals.deployment = false;
            await this.fetchData();
        },

        async deleteDeployment(ip) {
            let encodedIp = ip === '' ? '__all__' : encodeURIComponent(ip);
            if(confirm('Delete deployment for ' + (ip || 'all IPs') + '?')) {
                await fetch('/api/deployments/' + encodedIp, { method: 'DELETE' });
                await this.fetchData();
            }
        },

        openTemplateModal() {
            this.forms.template = { name: '', variables: [], content: '#!ipxe\n' };
            this.modals.template = true;
        },

        editTemplate(t) {
            this.forms.template = JSON.parse(JSON.stringify(t));
            this.modals.template = true;
        },

        async saveTemplate() {
            await fetch('/api/templates', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.forms.template)
            });
            this.modals.template = false;
            await this.fetchData();
        },

        async deleteTemplate(id) {
            if(confirm('Delete template?')) {
                await fetch('/api/templates/' + id, { method: 'DELETE' });
                await this.fetchData();
            }
        },

        async uploadAsset(e) {
            const file = e.target.files[0];
            if (!file) return;
            
            const formData = new FormData();
            formData.append('file', file);
            
            await fetch('/api/assets', {
                method: 'POST',
                body: formData
            });
            
            e.target.value = ''; // reset
            await this.fetchData();
        },

        openAssetTextModal() {
            this.forms.asset = { id: '', filename: '', content: '' };
            this.modals.asset = true;
        },

        async editAsset(a) {
            const res = await fetch('/a/' + a.id + '/' + a.filename);
            const content = await res.text();
            this.forms.asset = { id: a.id, filename: a.filename, content: content };
            this.modals.asset = true;
        },

        async saveAssetText() {
            await fetch('/api/assets/text', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.forms.asset)
            });
            this.modals.asset = false;
            await this.fetchData();
        },

        async deleteAsset(id) {
            if(confirm('Delete asset?')) {
                await fetch('/api/assets/' + id, { method: 'DELETE' });
                await this.fetchData();
            }
        }
    }
}
