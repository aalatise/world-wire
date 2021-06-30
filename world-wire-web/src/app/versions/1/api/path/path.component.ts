import { Component, OnInit } from '@angular/core';
import { SwaggerService } from '../../../shared/services/swagger.service';
import * as _ from 'lodash';

@Component({
  selector: 'app-path',
  templateUrl: './path.component.html',
  styleUrls: ['./path.component.scss']
})
export class PathComponent implements OnInit {
  constructor(
    public swaggerService: SwaggerService
  ) { }

  ngOnInit() { }

}
