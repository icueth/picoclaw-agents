import os
import glob
import re

source_dir = '/Users/icue/Downloads/agency-agents-main'

# Scan for all .md files in department folders
identity_files = glob.glob(f'{source_dir}/*/*.md', recursive=False)

agents = []
for file_path in identity_files:
    parts = file_path.split('/')
    if len(parts) >= 2:
        department = parts[-2]
        filename = parts[-1]
        if filename.lower() == 'readme.md' or filename.lower() == 'identity.md':
            continue
            
        slug = filename.replace('.md', '')
        if slug.startswith(department + '-'):
            slug = slug[len(department) + 1:]

        
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
            
        name = slug
        role = slug
        avatar = "🤖"
        description = "An agent."
        capabilities = "[]"
        
        name_match = re.search(r'name:\s*(.+)', content)
        if name_match: name = name_match.group(1).strip('\'"')
            
        role_match = re.search(r'role:\s*(.+)', content)
        if role_match: role = role_match.group(1).strip('\'"')
            
        avatar_match = re.search(r'avatar:\s*(.+)', content)
        if avatar_match: avatar = avatar_match.group(1).strip('\'"')
            
        desc_match = re.search(r'description:\s*(.+)', content)
        if desc_match: description = desc_match.group(1).strip('\'"')
            
        # Parse capabilities block
        cap_match = re.search(r'capabilities:\n((?:  - .+\n)+)', content)
        if cap_match:
            caps = re.findall(r'- (.+)', cap_match.group(1))
            capabilities = repr(caps).replace("'", '"')
            
        agents.append({
            'slug': slug,
            'name': name,
            'department': department,
            'role': role,
            'avatar': avatar,
            'description': description,
            'capabilities': capabilities,
            'content': content
        })

mapped_files = {
    'builtin_agents_core_eng.go': [('coreAgents', ['core']), ('engineeringAgents', ['engineering'])],
    'builtin_agents_design_mkt.go': [('designAgents', ['design']), ('marketingAgents', ['marketing'])],
    'builtin_agents_game_spatial_media.go': [('gameDevelopmentAgents', ['game-development']), ('spatialComputingAgents', ['spatial-computing']), ('paidMediaAgents', ['paid-media'])],
    'builtin_agents_specialized.go': [('specializedAgents', ['specialized', 'examples', 'integrations', 'scripts'])],
    'builtin_agents_test_prod_pm_sup.go': [('testingAgents', ['testing']), ('productAgents', ['product']), ('projectManagementAgents', ['project-management']), ('supportAgents', ['support']), ('strategyAgents', ['strategy'])]
}

for filename, func_mappings in mapped_files.items():
    file_path = f'/Users/icue/Code/AI/picoclaw/pkg/agent/{filename}'
    
    out = []
    out.append('package agent')
    out.append('')
    
    total_count = 0
    for func_name, depts in func_mappings:
        out.append(f'// {func_name} returns built-in agents.')
        out.append(f'func {func_name}() []BuiltinAgent {{')
        out.append('\treturn []BuiltinAgent{')
        
        count = 0
        for dept in depts:
            agents_for_dept = [a for a in agents if a['department'] == dept]
            for a in agents_for_dept:
                count += 1
                total_count += 1
                escaped_prompt = a['content'].replace('`', '`+"`"+`')
                desc_escaped = a['description'].replace('"', '\\"')
                caps_formatted = a['capabilities'].replace('["', '{"').replace('"]', '"}').replace("[]", "{}")
                
                out.append(f"""\t\t{{
\t\t\tID:             "{a['slug']}",
\t\t\tName:           "{a['name']}",
\t\t\tDepartment:     "{a['department']}",
\t\t\tRole:           "{a['role']}",
\t\t\tAvatar:         "{a['avatar']}",
\t\t\tDescription:    "{desc_escaped}",
\t\t\tCapabilities:   []string{caps_formatted},
\t\t\tIsPermanent:    false,
\t\t\tPrompt: `{escaped_prompt}`,
\t\t}},""")
        
        out.append('\t}')
        out.append('}')
        out.append('')
    
    if total_count > 0:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write('\n'.join(out) + '\n')
        print(f'Wrote {total_count} agents to {filename}')
    else:
        print(f'No agents found for {filename}')
