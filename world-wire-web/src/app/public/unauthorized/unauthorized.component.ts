import { Component, OnInit, HostBinding } from '@angular/core';

@Component({
  selector: 'app-unauthorized',
  templateUrl: './unauthorized.component.html',
  styleUrls: ['./unauthorized.component.scss']
})
export class UnauthorizedComponent implements OnInit {

  constructor() { }

  @HostBinding('attr.class') cls = 'flex-fill';

  ngOnInit() {
  }

}
