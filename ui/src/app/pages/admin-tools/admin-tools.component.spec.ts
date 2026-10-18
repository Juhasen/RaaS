import { TestBed } from '@angular/core/testing';
import { AdminToolsComponent } from './admin-tools.component';
import { INFRA_TOOLS } from '../../config/infra-tools';

describe('AdminToolsComponent', () => {
  it('opens each tool in a new tab', async () => {
    const fixture = TestBed.createComponent(AdminToolsComponent);
    await fixture.whenStable();
    const links = Array.from(fixture.nativeElement.querySelectorAll('a')) as HTMLAnchorElement[];

    expect(links.map((a) => a.getAttribute('href'))).toEqual([
      'http://localhost:30901',
      'http://localhost:30090',
      'http://localhost:30808',
      'http://localhost:30300'
    ]);
    expect(INFRA_TOOLS.map((t) => t.name)).toEqual(['MinIO', 'Postgres (Adminer)', 'Kafka (Redpanda Console)', 'Grafana']);
    for (const a of links) {
      expect(a.target).toBe('_blank');
      expect(a.rel).toBe('noopener noreferrer');
    }
  });
});
