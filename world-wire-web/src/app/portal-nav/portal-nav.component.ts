import { Component, OnInit, AfterViewInit, ChangeDetectorRef, ViewChild, ElementRef } from '@angular/core';
import { AuthService } from '../shared/services/auth.service';

@Component({
  selector: 'app-portal-nav',
  templateUrl: './portal-nav.component.html',
  styleUrls: ['./portal-nav.component.scss']
})
export class PortalNavComponent implements OnInit {

  sidenav: any; // prevents linting error

  @ViewChild('mainMenu') mainMenu: ElementRef;

  showMenu = false;

  constructor(
    public authService: AuthService
  ) {
    this.sidenav = '';
  }

  ngOnInit() { }

}
