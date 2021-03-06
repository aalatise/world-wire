import { Component, OnInit } from '@angular/core';
import { Tooltip } from 'carbon-components';
import { VersionService } from '../../shared/services/version.service';


@Component({
  selector: 'app-guides',
  templateUrl: './guides.component.html',
  styleUrls: ['./guides.component.scss']
})
export class GuidesComponent implements OnInit {
  constructor(
    public versionService: VersionService
  ) { }

  ngOnInit() {
  }

}
