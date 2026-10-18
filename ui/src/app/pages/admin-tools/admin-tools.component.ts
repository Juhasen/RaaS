import { ChangeDetectionStrategy, Component } from '@angular/core';
import { INFRA_TOOLS } from '../../config/infra-tools';

@Component({
  selector: 'app-admin-tools',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <h1 class="text-2xl font-semibold tracking-tight text-zinc-900 mb-1">Infrastructure tools</h1>
    <p class="text-sm text-zinc-500 mb-6">Each tool opens in a new tab and has its own login.</p>
    <ul class="grid gap-3 sm:grid-cols-2">
      @for (tool of tools; track tool.id) {
        <li>
          <a
            [href]="tool.url"
            target="_blank"
            rel="noopener noreferrer"
            class="block h-full p-4 bg-white border border-zinc-200 rounded-xl hover:border-indigo-300 hover:shadow-md transition-all focus-visible:outline-2 focus-visible:outline-indigo-600"
          >
            <span class="block font-semibold text-zinc-900">{{ tool.name }}</span>
            <span class="block text-sm text-zinc-600">{{ tool.description }}</span>
          </a>
        </li>
      }
    </ul>
  `
})
export class AdminToolsComponent {
  readonly tools = INFRA_TOOLS;
}
