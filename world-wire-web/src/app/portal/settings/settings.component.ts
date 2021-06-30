import { Component, OnInit, HostBinding } from '@angular/core';

@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {

  constructor() { }

  @HostBinding('attr.class') cls = 'flex-fill';

  ngOnInit() {
  }

}
