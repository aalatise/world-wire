import { Component, OnInit } from '@angular/core';
import { AuthService } from '../shared/services/auth.service';

@Component({
  selector: 'app-public-nav',
  templateUrl: './public-nav.component.html',
  styleUrls: ['./public-nav.component.scss']
})
export class PublicNavComponent implements OnInit {

  sidenav: any;

  constructor(
    public auth: AuthService
    ) {
    this.sidenav = ''; // prevents linting error
  }

  ngOnInit() {
  }

}
